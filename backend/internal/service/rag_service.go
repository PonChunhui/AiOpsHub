package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

// KnowledgeDocument 知识库文档结构
// 用于存储和管理运维知识库中的文档数据
type KnowledgeDocument struct {
	ID        string                 `json:"id"`        // 文档唯一标识符
	Title     string                 `json:"title"`     // 文档标题
	Content   string                 `json:"content"`   // 文档内容（支持Markdown格式）
	DocType   string                 `json:"doc_type"`  // 文档类型：sop / faq / alert
	Component string                 `json:"component"` // 组件名：mysql / k8s / redis
	Tags      []string               `json:"tags"`      // 文档标签（用于分类和检索）
	Metadata  map[string]interface{} `json:"metadata"`  // 文档元数据（创建时间、作者等）
}

// SearchResult 知识检索结果结构
// 包含检索到的文档及其相关性评分
type SearchResult struct {
	Document       KnowledgeDocument `json:"document"`
	Score          float64           `json:"score"`
	Distance       float64           `json:"distance"`
	RelevanceLevel string            `json:"relevance_level"`
}

// RAGService RAG（检索增强生成）服务
// 提供知识库检索功能，支持向量检索和内存检索两种模式
type RAGService struct {
	Collection    string
	Enabled       bool
	MilvusService *MilvusService
	EmbeddingSvc  *EmbeddingService
	DBRepo        *repository.RAGRepository
}

// NewRAGService 创建RAG服务（内存模式）
// 适用于Milvus不可用时的降级方案，使用内置的硬编码知识库
// 参数：collection - collection名称（用于标识）
// 返回：RAG服务实例（仅支持内存检索）
func NewRAGService(collection string) *RAGService {
	service := &RAGService{
		Collection: collection,
		Enabled:    true,
		DBRepo:     repository.NewRAGRepository(),
	}

	logger.Info("RAG Service created (memory mode with PostgreSQL metadata storage)")
	return service
}

func NewRAGServiceWithMilvus(collection string, milvusSvc *MilvusService, embeddingSvc *EmbeddingService) *RAGService {
	service := &RAGService{
		Collection:    collection,
		Enabled:       true,
		MilvusService: milvusSvc,
		EmbeddingSvc:  embeddingSvc,
		DBRepo:        repository.NewRAGRepository(),
	}

	logger.Info("RAG Service created with Milvus backend and PostgreSQL metadata storage")
	return service
}

func (r *RAGService) SearchKnowledge(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	logger.Debug(fmt.Sprintf("Searching knowledge: query='%s', topK=%d", query, topK))

	const (
		HighRelevanceThreshold   = 0.70
		MediumRelevanceThreshold = 0.50
		LowRelevanceThreshold    = 0.40
	)

	// 优先使用Milvus向量检索（推荐方式）
	if r.MilvusService != nil && r.EmbeddingSvc != nil {
		// 将查询文本转换为向量
		queryEmbedding, err := r.EmbeddingSvc.GetEmbedding(ctx, query)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get query embedding: %v", err))
			return nil, err
		}

		// 在Milvus中进行向量相似度检索
		results, err := r.MilvusService.SearchDocuments(ctx, queryEmbedding, topK)
		if err != nil {
			return nil, err
		}

		// 过滤低相关性结果，使用混合检索策略（Keyword + Vector）
		// 混合检索是业界主流方案，结合关键词匹配和语义相似度的优势
		// Keyword权重20%，Vector权重80%
		filteredResults := []SearchResult{}
		for _, result := range results {
			// 计算关键词匹配分数
			keywordScore := calculateKeywordScore(query, result.Document.Title, result.Document.Content)

			// 混合评分：Vector * 0.8 + Keyword * 0.2
			finalScore := result.Score*0.8 + keywordScore*0.2

			var relevanceLevel string
			var included bool

			if finalScore >= HighRelevanceThreshold {
				relevanceLevel = "high"
				included = true
			} else if finalScore >= MediumRelevanceThreshold {
				relevanceLevel = "medium"
				included = true
			} else if finalScore >= LowRelevanceThreshold {
				relevanceLevel = "low"
				included = false
				logger.Debug(fmt.Sprintf("Filtered low relevance document: title='%s', keyword=%.4f, vector=%.4f, final=%.4f",
					result.Document.Title, keywordScore, result.Score, finalScore))
			} else {
				relevanceLevel = "none"
				included = false
				logger.Debug(fmt.Sprintf("Filtered irrelevant document: title='%s', keyword=%.4f, vector=%.4f, final=%.4f",
					result.Document.Title, keywordScore, result.Score, finalScore))
			}

			logger.Debug(fmt.Sprintf("RAG search result: title='%s', keyword=%.4f, vector=%.4f, final=%.4f, level=%s, included=%v",
				result.Document.Title, keywordScore, result.Score, finalScore, relevanceLevel, included))

			if included {
				filteredResults = append(filteredResults, SearchResult{
					Document:       result.Document,
					Score:          finalScore,
					Distance:       result.Distance,
					RelevanceLevel: relevanceLevel,
				})
			}
		}

		if len(filteredResults) == 0 {
			logger.Debug(fmt.Sprintf("No relevant knowledge found for query='%s' (threshold=%.4f)",
				query, MediumRelevanceThreshold))
		} else {
			logger.Info(fmt.Sprintf("Found %d relevant results from Milvus (filtered from %d, threshold=%.4f)",
				len(filteredResults), len(results), MediumRelevanceThreshold))
		}
		return filteredResults, nil
	}

	// 降级到内存检索（兜底方案）
	// 内存检索使用阈值0.75，过滤相关性较低的文档
	results := []SearchResult{}
	docs := r.searchInMemory(query, topK)

	for _, doc := range docs {
		score := r.calculateScore(query, doc.Content)
		var relevanceLevel string
		var included bool

		if score >= HighRelevanceThreshold {
			relevanceLevel = "high"
			included = true
		} else if score >= MediumRelevanceThreshold {
			relevanceLevel = "medium"
			included = true
		} else if score >= LowRelevanceThreshold {
			relevanceLevel = "low"
			included = false
			logger.Debug(fmt.Sprintf("Filtered low relevance memory document: title=%s, score=%.4f",
				doc.Title, score))
		} else {
			relevanceLevel = "none"
			included = false
			logger.Debug(fmt.Sprintf("Filtered irrelevant memory document: title=%s, score=%.4f",
				doc.Title, score))
		}

		logger.Debug(fmt.Sprintf("Memory search result: title=%s, score=%.4f, level=%s, included=%v",
			doc.Title, score, relevanceLevel, included))

		if included {
			results = append(results, SearchResult{
				Document:       doc,
				Score:          score,
				Distance:       1 - score,
				RelevanceLevel: relevanceLevel,
			})
		}
	}

	if len(results) == 0 {
		logger.Debug(fmt.Sprintf("No relevant knowledge found in memory for query='%s' (threshold=%.4f)",
			query, MediumRelevanceThreshold))
	} else {
		logger.Info(fmt.Sprintf("Found %d relevant results from memory (threshold=%.4f)", len(results), MediumRelevanceThreshold))
	}
	return results, nil
}

// searchInMemory 内存知识库检索（降级/兜底方案）
// 当Milvus向量数据库不可用时，使用内置的硬编码知识库进行检索
//
// 重要说明：
// - 此方法仅包含5条硬编码的示例运维文档（kb-001到kb-005）
// - 无法检索用户通过API添加到Milvus的知识库文档
// - 仅作为Milvus不可用时的兜底方案，确保RAG功能不会完全失效
//
// 使用场景：
// 1. Milvus服务启动失败或连接异常
// 2. Embedding服务不可用，无法生成向量
// 3. 开发/测试环境，无需完整的向量检索功能
//
// 检索方式：
// - 使用简单的关键词匹配（matchQuery方法）
// - 不支持语义相似度检索
// - 检索效果不如向量检索准确
//
// 参数：query - 查询内容（关键词）
//
//	limit - 返回结果数量限制
//
// 返回：匹配的知识库文档列表（最多limit条）
func (r *RAGService) searchInMemory(query string, limit int) []KnowledgeDocument {
	// 内置的示例知识库（5条运维文档）
	// 注意：这些文档是硬编码的，用户添加的文档不会被检索
	knowledgeBase := []KnowledgeDocument{
		{
			ID:        "kb-001",
			Title:     "服务响应慢常见原因",
			Content:   "# 服务响应慢常见原因\n\n## CPU相关问题\n- CPU使用率过高\n- CPU throttling\n- 进程死循环\n\n## 内存相关问题\n- 内存不足\n- 内存泄漏\n- 大对象未释放\n\n## 数据库问题\n- 慢查询\n- 索引缺失\n- 连接池配置不当\n\n## 解决方案\n1. 添加索引\n2. 优化SQL\n3. 使用缓存",
			DocType:   "sop",
			Component: "general",
			Tags:      []string{"性能", "响应慢", "故障排查"},
		},
		{
			ID:        "kb-002",
			Title:     "CPU使用率高排查方法",
			Content:   "# CPU使用率高排查\n\n## 排查步骤\n```bash\n# 1. 查看CPU使用情况\ntop -p <pid>\n\n# 2. 分析线程状态\nps -Lp <pid>\n\n# 3. 查看线程CPU占用\ntop -H -p <pid>\n```\n\n## 常见原因\n1. 死循环\n2. 频繁GC\n3. 算法复杂度高\n4. 锁竞争激烈",
			DocType:   "sop",
			Component: "general",
			Tags:      []string{"CPU", "性能排查"},
		},
		{
			ID:        "kb-003",
			Title:     "数据库慢查询优化",
			Content:   "# 数据库慢查询优化\n\n## 优化方法\n\n### 1. 添加索引\n```sql\nCREATE INDEX idx_user_id ON users(user_id);\n```\n\n### 2. 优化SQL语句\n- 避免 SELECT *\n- 使用 JOIN 替代子查询\n- 合理使用 LIMIT\n\n### 3. 分库分表\n- 水平分表\n- 垂直分库\n\n### 4. 使用缓存\n- Redis缓存热点数据\n- 本地缓存配置信息",
			DocType:   "sop",
			Component: "mysql",
			Tags:      []string{"数据库", "优化", "慢查询"},
		},
		{
			ID:        "kb-004",
			Title:     "内存泄漏检测",
			Content:   "# 内存泄漏检测\n\n## 检测方法\n\n### Go程序\n```bash\n# pprof分析\ngo tool pprof http://localhost:6060/debug/pprof/heap\n\n# 查看内存趋势\ncurl http://localhost:6060/debug/pprof/heap > heap.out\n```\n\n### Java程序\n```bash\n# jmap查看堆内存\njmap -histo <pid>\n\n# MAT分析堆转储\njmap -dump:format=b,file=heap.hprof <pid>\n```",
			DocType:   "sop",
			Component: "general",
			Tags:      []string{"内存", "泄漏", "排查"},
		},
		{
			ID:        "kb-005",
			Title:     "Kubernetes资源限制配置",
			Content:   "# Kubernetes资源限制配置\n\n## 资源配置示例\n```yaml\nresources:\n  requests:\n    cpu: \"100m\"\n    memory: \"128Mi\"\n  limits:\n    cpu: \"500m\"\n    memory: \"512Mi\"\n```\n\n## 最佳实践\n- requests应设置为正常负载下的需求\n- limits应设置为峰值负载下的上限\n- 避免limits远大于requests（避免资源浪费）",
			DocType:   "sop",
			Component: "k8s",
			Tags:      []string{"K8s", "资源", "配置"},
		},
	}

	// 遍历知识库，通过关键词匹配检索相关文档
	var results []KnowledgeDocument
	for _, doc := range knowledgeBase {
		// 使用matchQuery方法进行关键词匹配
		if r.matchQuery(query, doc) {
			results = append(results, doc)
			// 达到数量限制后停止检索
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// matchQuery 关键词匹配查询（用于内存检索）
// 通过关键词匹配判断文档是否与查询相关
//
// 改进的匹配策略（避免误匹配）：
// 1. 精确匹配：查询词与文档标签或分类完全匹配（高相关性）
// 2. 多关键词匹配：查询词包含≥2个关键词，且文档内容也包含这些关键词（中等相关性）
// 3. 单关键词匹配已禁用：避免误匹配边缘相关内容
//
// 预定义关键词列表：慢、性能、CPU、内存、数据库、优化、排查、故障、响应、K8s、Kubernetes
//
// 注意：此方法仅用于内存检索，不支持语义理解
//
//	向量检索（Milvus）效果更好，支持语义相似度匹配
//
// 参数：query - 查询内容
//
//	doc - 待匹配的知识库文档
//
// 返回：是否匹配（true表示相关）
func (r *RAGService) matchQuery(query string, doc KnowledgeDocument) bool {
	queryLower := strings.ToLower(query)

	// 精确匹配：查询词与标签完全匹配（相关性最高）
	for _, tag := range doc.Tags {
		if strings.ToLower(queryLower) == strings.ToLower(tag) {
			return true
		}
	}

	// 精确匹配：查询词与分类完全匹配（相关性高）
	if strings.Contains(queryLower, strings.ToLower(doc.DocType)) || strings.Contains(queryLower, strings.ToLower(doc.Component)) {
		return true
	}

	// 多关键词匹配策略（避免单关键词误匹配）
	// 预定义关键词列表（运维常见问题关键词）
	keywords := []string{"慢", "性能", "CPU", "内存", "数据库", "优化", "排查", "故障", "响应", "K8s", "Kubernetes"}

	// 计算查询词包含的关键词数量
	queryKeywordCount := 0
	for _, keyword := range keywords {
		if strings.Contains(queryLower, keyword) {
			queryKeywordCount++
		}
	}

	// 查询词必须包含≥2个关键词才进行匹配（避免单关键词误匹配）
	if queryKeywordCount >= 2 {
		// 检查文档是否也包含这些关键词
		docKeywordMatch := 0
		for _, keyword := range keywords {
			if strings.Contains(queryLower, keyword) {
				// 文档标签包含关键词
				for _, tag := range doc.Tags {
					if strings.Contains(strings.ToLower(tag), keyword) {
						docKeywordMatch++
						break
					}
				}
				// 文档分类包含关键词
				if strings.Contains(strings.ToLower(doc.DocType), keyword) || strings.Contains(strings.ToLower(doc.Component), keyword) {
					docKeywordMatch++
				}
				// 文档内容包含关键词
				if strings.Contains(strings.ToLower(doc.Content), keyword) {
					docKeywordMatch++
				}
			}
		}

		// 文档必须匹配至少2个查询关键词才认为相关
		if docKeywordMatch >= 2 {
			return true
		}
	}

	return false
}

// calculateScore 计算相关性评分（用于内存检索）
// 为内存检索的匹配结果计算相关性评分，基于关键词匹配程度动态评分
//
// 评分策略：
// 1. 精确匹配（标签或分类完全匹配）：评分0.90（最高）
// 2. 关键词多词匹配：评分0.75-0.85（根据匹配关键词数量）
// 3. 单关键词匹配：评分0.70（较低）
//
// 参数：query - 查询内容
//
//	content - 文档内容
//
// 返回：相关性评分（0.70-0.90）
func (r *RAGService) calculateScore(query, content string) float64 {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	// 计算匹配的关键词数量（用于评估相关性强度）
	keywords := []string{"慢", "性能", "CPU", "内存", "数据库", "优化", "排查", "故障", "响应", "K8s", "Kubernetes"}
	matchedKeywords := 0
	for _, keyword := range keywords {
		if strings.Contains(queryLower, keyword) && strings.Contains(contentLower, keyword) {
			matchedKeywords++
		}
	}

	// 根据匹配关键词数量计算评分
	if matchedKeywords >= 3 {
		// 多关键词匹配，相关性高
		return 0.85
	} else if matchedKeywords >= 2 {
		// 中等关键词匹配
		return 0.80
	} else if matchedKeywords >= 1 {
		// 单关键词匹配，相关性较低
		return 0.70
	}

	// 没有关键词匹配（可能是标签精确匹配）
	// 返回基础评分，需要进一步通过阈值过滤
	return 0.75
}

// AddDocument 添加知识库文档
// 将文档添加到Milvus向量数据库，使其可以通过向量检索被搜索到
//
// 工作流程：
// 1. 如果Milvus和Embedding服务可用：
//   - 使用Embedding服务将文档内容转换为向量
//   - 将文档和向量一起插入到Milvus
//   - 文档可以被向量检索搜索到（推荐）
//
// 2. 如果Milvus不可用：
//   - 不执行任何操作（无法添加到内存知识库）
//   - 文档无法被检索到
//
// 重要说明：
// - 文档只会添加到Milvus，不会添加到内存知识库
// - 内存知识库的5条文档是硬编码的，无法动态添加
// - 只有添加到Milvus的文档才能被向量检索搜索到
//
// 参数：ctx - 上下文
//
//	doc - 知识库文档（包含ID、标题、内容、分类、标签等）
//
// 返回：错误信息（nil表示成功）
func (r *RAGService) AddDocument(ctx context.Context, doc KnowledgeDocument) error {
	logger.Debug(fmt.Sprintf("Adding document: %s (%s)", doc.ID, doc.Title))

	tagsJSON, _ := json.Marshal(doc.Tags)

	createdAt := time.Now()
	createdBy := ""
	if doc.Metadata != nil {
		if v, ok := doc.Metadata["created_by"]; ok {
			createdBy = fmt.Sprintf("%v", v)
		}
	}

	pgDoc := &model.RAGDocument{
		ID:        doc.ID,
		Title:     doc.Title,
		Content:   doc.Content,
		DocType:   doc.DocType,
		Component: doc.Component,
		Tags:      string(tagsJSON),
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	if r.DBRepo != nil {
		if err := r.DBRepo.Create(pgDoc); err != nil {
			logger.Error(fmt.Sprintf("Failed to save document to PostgreSQL: %v", err))
			return err
		}
		logger.Debug(fmt.Sprintf("Document saved to PostgreSQL: %s", doc.ID))
	}

	if r.MilvusService != nil && r.EmbeddingSvc != nil {
		embedding, err := r.EmbeddingSvc.GetEmbedding(ctx, doc.Content)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get document embedding: %v", err))
			return err
		}

		err = r.MilvusService.InsertDocument(ctx, doc, embedding)
		if err != nil {
			return err
		}

		logger.Debug(fmt.Sprintf("Document added to Milvus: %s", doc.ID))
	}

	return nil
}

// GetDocument 获取知识库文档
// 根据文档ID从Milvus或内存知识库中获取文档详情
//
// 查询优先级：
// 1. 如果Milvus可用：从Milvus查询（包含用户添加的所有文档）
// 2. 如果Milvus不可用：从内存知识库查询（仅包含5条硬编码文档）
//
// 参数：ctx - 上下文
//
//	docID - 文档ID
//
// 返回：文档内容和错误信息
func (r *RAGService) GetDocument(ctx context.Context, docID string) (*KnowledgeDocument, error) {
	logger.Debug(fmt.Sprintf("Getting document: %s", docID))

	// 优先从Milvus获取
	if r.MilvusService != nil {
		return r.MilvusService.GetDocument(ctx, docID)
	}

	// 降级到内存知识库（仅包含硬编码的5条文档）
	docs := r.searchInMemory(docID, 1)
	if len(docs) > 0 {
		return &docs[0], nil
	}

	return nil, fmt.Errorf("document not found: %s", docID)
}

// UpdateDocument 更新知识库文档
// 更新文档内容、标题、分类等信息，并重新生成向量
//
// 更新流程（Milvus模式）：
// 1. 删除旧文档（DeleteDocument）
// 2. 使用新内容生成向量（GetEmbedding）
// 3. 插入更新后的文档（InsertDocument）
//
// 注意：Milvus不支持直接更新，需要先删除再插入
//
// 参数：ctx - 上下文
//
//	docID - 文档ID
//	title - 新标题
//	content - 新内容
//	category - 新分类
//	tags - 新标签
//	metadata - 元数据（包含更新时间等）
//
// 返回：更新后的文档和错误信息
func (r *RAGService) UpdateDocument(ctx context.Context, docID string, title string, content string, docType string, component string, tags []string, metadata map[string]interface{}) (*KnowledgeDocument, error) {
	logger.Debug(fmt.Sprintf("Updating document: %s", docID))

	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["updated_at"] = time.Now().Format(time.RFC3339)

	doc := &KnowledgeDocument{
		ID:        docID,
		Title:     title,
		Content:   content,
		DocType:   docType,
		Component: component,
		Tags:      tags,
		Metadata:  metadata,
	}

	tagsJSON, _ := json.Marshal(tags)
	updatedBy := ""
	if v, ok := metadata["updated_by"]; ok {
		updatedBy = fmt.Sprintf("%v", v)
	}

	if r.DBRepo != nil {
		pgDoc, err := r.DBRepo.GetByID(docID)
		if err == nil {
			pgDoc.Title = title
			pgDoc.Content = content
			pgDoc.DocType = docType
			pgDoc.Component = component
			pgDoc.Tags = string(tagsJSON)
			pgDoc.UpdatedBy = updatedBy
			pgDoc.UpdatedAt = time.Now()
			if err := r.DBRepo.Update(pgDoc); err != nil {
				logger.Error(fmt.Sprintf("Failed to update document in PostgreSQL: %v", err))
			} else {
				logger.Debug(fmt.Sprintf("Document updated in PostgreSQL: %s", docID))
			}
		}
	}

	if r.MilvusService != nil && r.EmbeddingSvc != nil {
		err := r.MilvusService.DeleteDocument(ctx, docID)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to delete old document: %v", err))
		}

		embedding, err := r.EmbeddingSvc.GetEmbedding(ctx, content)
		if err != nil {
			return nil, err
		}

		err = r.MilvusService.InsertDocument(ctx, *doc, embedding)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

func (r *RAGService) DeleteDocument(ctx context.Context, docID string) error {
	logger.Debug(fmt.Sprintf("Deleting document: %s", docID))

	if r.DBRepo != nil {
		if err := r.DBRepo.Delete(docID); err != nil {
			logger.Error(fmt.Sprintf("Failed to delete document from PostgreSQL: %v", err))
		} else {
			logger.Debug(fmt.Sprintf("Document deleted from PostgreSQL: %s", docID))
		}
	}

	if r.MilvusService != nil {
		return r.MilvusService.DeleteDocument(ctx, docID)
	}

	return nil
}

func (r *RAGService) ListDocuments(ctx context.Context, docType string, component string, search string, page, pageSize int) ([]KnowledgeDocument, int64, error) {
	logger.Debug(fmt.Sprintf("Listing documents: docType=%s, component=%s, search=%s, page=%d, pageSize=%d", docType, component, search, page, pageSize))

	if r.DBRepo != nil {
		pgDocs, total, err := r.DBRepo.List(docType, component, search, page, pageSize)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to list documents from PostgreSQL: %v", err))
			return nil, 0, err
		}

		docs := []KnowledgeDocument{}
		for _, pgDoc := range pgDocs {
			var tags []string
			if pgDoc.Tags != "" {
				json.Unmarshal([]byte(pgDoc.Tags), &tags)
			}

			doc := KnowledgeDocument{
				ID:        pgDoc.ID,
				Title:     pgDoc.Title,
				Content:   pgDoc.Content,
				DocType:   pgDoc.DocType,
				Component: pgDoc.Component,
				Tags:      tags,
				Metadata: map[string]interface{}{
					"created_at": pgDoc.CreatedAt.Format(time.RFC3339),
					"created_by": pgDoc.CreatedBy,
					"updated_at": pgDoc.UpdatedAt.Format(time.RFC3339),
					"updated_by": pgDoc.UpdatedBy,
				},
			}
			docs = append(docs, doc)
		}

		logger.Debug(fmt.Sprintf("Listed %d documents from PostgreSQL (total: %d)", len(docs), total))
		return docs, total, nil
	}

	if r.MilvusService != nil {
		docs, err := r.MilvusService.ListDocuments(ctx, 10000)
		if err != nil {
			return nil, 0, err
		}
		return docs, int64(len(docs)), nil
	}

	return r.searchInMemory(search, pageSize), int64(len(r.searchInMemory(search, pageSize))), nil
}

func (r *RAGService) GetContextForQuery(ctx context.Context, query string, maxTokens int) (string, error) {
	results, err := r.SearchKnowledge(ctx, query, 3)
	if err != nil {
		return "", err
	}

	context := "相关知识背景:\n"
	for i, result := range results {
		context += fmt.Sprintf("%d. %s\n   %s\n\n", i+1, result.Document.Title, result.Document.Content)
	}

	logger.Debug(fmt.Sprintf("Generated context: %d chars", len(context)))
	return context, nil
}

// calculateKeywordScore 使用关键词匹配计算分数
// 适用于中英文混合场景，无需第三方库
func calculateKeywordScore(query, title, content string) float64 {
	// 提取查询关键词（停用词过滤）
	keywords := extractKeywords(query)

	if len(keywords) == 0 {
		return 0.0
	}

	// 合并标题和内容作为文档
	document := strings.ToLower(title + " " + content)

	// 计算关键词匹配比例
	matchedCount := 0
	for _, keyword := range keywords {
		if strings.Contains(document, keyword) {
			matchedCount++
		}
	}

	matchRatio := float64(matchedCount) / float64(len(keywords))
	return matchRatio // 0.0-1.0
}

// extractKeywords 提取关键词（停用词过滤）
func extractKeywords(text string) []string {
	textLower := strings.ToLower(text)

	// 常见停用词（中英文）
	stopWords := map[string]bool{
		"怎么": true, "如何": true, "怎样": true, "为什么": true,
		"什么": true, "哪": true, "哪里": true, "哪个": true,
		"请": true, "帮": true, "帮忙": true, "能否": true,
		"能够": true, "可以": true, "能": true, "用": true,
		"让": true, "给": true, "的": true, "了": true,
		"在": true, "是": true, "和": true, "与": true,
		"或": true, "有": true, "这": true, "那": true,
		"the": true, "a": true, "an": true, "is": true,
		"are": true, "was": true, "how": true, "what": true,
		"why": true, "where": true, "which": true, "can": true,
		"could": true, "would": true, "please": true, "help": true,
		"need": true, "want": true, "do": true, "did": true,
		"in": true, "on": true, "at": true, "to": true, "of": true,
		"for": true, "with": true, "by": true, "from": true,
	}

	keywords := []string{}

	// 1. 提取英文单词和数字组合（如kubekey, k8s, wifi等）
	englishPattern := regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9\-]+`)
	englishMatches := englishPattern.FindAllString(textLower, -1)
	for _, match := range englishMatches {
		if !stopWords[match] {
			keywords = append(keywords, match)
		}
	}

	// 2. 提取中文字符串（>=2字的非停用词片段）
	// 使用简单的策略：提取连续的中文字符
	chinesePattern := regexp.MustCompile(`[\p{Han}]+`)
	chineseMatches := chinesePattern.FindAllString(textLower, -1)
	for _, match := range chineseMatches {
		if len(match) >= 2 && !stopWords[match] {
			keywords = append(keywords, match)
		}
	}

	return keywords
}
