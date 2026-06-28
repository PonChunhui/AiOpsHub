package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusConfig struct {
	Host     string
	Port     string
	Database string
}

type MilvusService struct {
	client     client.Client
	config     MilvusConfig
	collection string
	dimension  int64
}

func NewMilvusService(host, port, database string) (*MilvusService, error) {
	config := MilvusConfig{
		Host:     host,
		Port:     port,
		Database: database,
	}

	milvusClient, err := client.NewClient(context.Background(), client.Config{
		Address: fmt.Sprintf("%s:%s", host, port),
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Milvus: %v", err))
		return nil, err
	}

	service := &MilvusService{
		client:     milvusClient,
		config:     config,
		collection: "knowledge_documents",
		dimension:  1536,
	}

	logger.Info(fmt.Sprintf("Milvus client connected: %s:%s", host, port))
	return service, nil
}

func (m *MilvusService) CreateCollection(ctx context.Context) error {
	has, err := m.client.HasCollection(ctx, m.collection)
	if err != nil {
		return err
	}

	// 如果collection已存在，直接加载，不删除重建（避免数据丢失）
	if has {
		logger.Info(fmt.Sprintf("Collection %s already exists, skipping creation to preserve data", m.collection))
		return nil
	}

	// 只在不存在时创建新collection
	schema := &entity.Schema{
		CollectionName: m.collection,
		Description:    "Knowledge documents for AIOps",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "title",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "500"},
			},
			{
				Name:       "content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50000"},
			},
			{
				Name:       "category",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "tags",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "500"},
			},
			{
				Name:       "embedding",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{"dim": fmt.Sprintf("%d", m.dimension)},
			},
			{
				Name:       "created_at",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "created_by",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "updated_at",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "updated_by",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
		},
	}

	err = m.client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return err
	}

	index, err := entity.NewIndexIvfFlat(entity.L2, 128)
	if err != nil {
		return err
	}

	err = m.client.CreateIndex(ctx, m.collection, "embedding", index, false)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Collection %s created successfully with content max_length=50000", m.collection))
	return nil
}

func (m *MilvusService) InsertDocument(ctx context.Context, doc KnowledgeDocument, embedding []float32) error {
	tagsJSON, _ := json.Marshal(doc.Tags)

	createdAt := ""
	if doc.Metadata != nil {
		if v, ok := doc.Metadata["created_at"]; ok {
			createdAt = fmt.Sprintf("%v", v)
		}
	}
	if createdAt == "" {
		createdAt = time.Now().Format(time.RFC3339)
	}

	createdBy := ""
	if doc.Metadata != nil {
		if v, ok := doc.Metadata["created_by"]; ok {
			createdBy = fmt.Sprintf("%v", v)
		}
	}

	updatedAt := ""
	if doc.Metadata != nil {
		if v, ok := doc.Metadata["updated_at"]; ok {
			updatedAt = fmt.Sprintf("%v", v)
		}
	}
	if updatedAt == "" {
		updatedAt = createdAt
	}

	updatedBy := ""
	if doc.Metadata != nil {
		if v, ok := doc.Metadata["updated_by"]; ok {
			updatedBy = fmt.Sprintf("%v", v)
		}
	}
	if updatedBy == "" {
		updatedBy = createdBy
	}

	columns := []entity.Column{
		entity.NewColumnVarChar("id", []string{doc.ID}),
		entity.NewColumnVarChar("title", []string{doc.Title}),
		entity.NewColumnVarChar("content", []string{doc.Content}),
		entity.NewColumnVarChar("category", []string{doc.Category}),
		entity.NewColumnVarChar("tags", []string{string(tagsJSON)}),
		entity.NewColumnFloatVector("embedding", int(m.dimension), [][]float32{embedding}),
		entity.NewColumnVarChar("created_at", []string{createdAt}),
		entity.NewColumnVarChar("created_by", []string{createdBy}),
		entity.NewColumnVarChar("updated_at", []string{updatedAt}),
		entity.NewColumnVarChar("updated_by", []string{updatedBy}),
	}

	_, err := m.client.Insert(ctx, m.collection, "", columns...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to insert document: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Document inserted: %s", doc.ID))
	return nil
}

func (m *MilvusService) SearchDocuments(ctx context.Context, queryEmbedding []float32, topK int) ([]SearchResult, error) {
	sp, err := entity.NewIndexIvfFlatSearchParam(16)
	if err != nil {
		return nil, err
	}

	results, err := m.client.Search(
		ctx,
		m.collection,
		[]string{},
		"",
		[]string{"id", "title", "content", "category", "tags", "created_at", "created_by", "updated_at", "updated_by"},
		[]entity.Vector{entity.FloatVector(queryEmbedding)},
		"embedding",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to search: %v", err))
		return nil, err
	}

	searchResults := []SearchResult{}
	for _, result := range results {
		for i := 0; i < result.ResultCount; i++ {
			id, _ := result.Fields.GetColumn("id").Get(i)
			title, _ := result.Fields.GetColumn("title").Get(i)
			content, _ := result.Fields.GetColumn("content").Get(i)
			category, _ := result.Fields.GetColumn("category").Get(i)
			tagsStr, _ := result.Fields.GetColumn("tags").Get(i)
			createdAt, _ := result.Fields.GetColumn("created_at").Get(i)
			createdBy, _ := result.Fields.GetColumn("created_by").Get(i)
			updatedAt, _ := result.Fields.GetColumn("updated_at").Get(i)
			updatedBy, _ := result.Fields.GetColumn("updated_by").Get(i)

			var tags []string
			if tagsStr != nil {
				json.Unmarshal([]byte(tagsStr.(string)), &tags)
			}

			metadata := map[string]interface{}{}
			if createdAt != nil {
				metadata["created_at"] = createdAt.(string)
			}
			if createdBy != nil {
				metadata["created_by"] = createdBy.(string)
			}
			if updatedAt != nil {
				metadata["updated_at"] = updatedAt.(string)
			}
			if updatedBy != nil {
				metadata["updated_by"] = updatedBy.(string)
			}

			doc := KnowledgeDocument{
				ID:       id.(string),
				Title:    title.(string),
				Content:  content.(string),
				Category: category.(string),
				Tags:     tags,
				Metadata: metadata,
			}

			// 计算相关性评分（修正的评分计算）
			// L2距离越小越相似，转换为0-1的评分
			// 实际向量距离范围：0-200（根据日志观察）
			// 使用指数衰减公式：score = exp(-distance/30)
			// 衰减系数30使评分对距离敏感度适中：
			//   distance=0 → score=1.0（完全匹配）
			//   distance=1.5 → score=0.95（高相关，阈值线）
			//   distance=3 → score=0.90（中等相关）
			//   distance=5 → score=0.85（边缘相关）
			//   distance=10 → score=0.72（弱相关，被过滤）
			//   distance=30 → score=0.37（不相关）
			//   distance=100 → score=0.04（完全不相关）
			// 注意：距离<2通常表示语义相似，距离>10表示不相关
			distance := float64(result.Scores[i])
			score := math.Exp(-distance / 30.0)

			// 确保评分在0-1范围内
			if score > 1.0 {
				score = 1.0
			}
			if score < 0 {
				score = 0
			}

			// 详细日志：记录距离和评分的关系
			logger.Info(fmt.Sprintf("Milvus search: title=%s, distance=%.4f, score=%.4f",
				doc.Title, distance, score))

			searchResults = append(searchResults, SearchResult{
				Document: doc,
				Score:    score,
				Distance: distance,
			})
		}
	}

	logger.Info(fmt.Sprintf("Found %d documents", len(searchResults)))
	return searchResults, nil
}

func (m *MilvusService) ListDocuments(ctx context.Context, limit int) ([]KnowledgeDocument, error) {
	if limit <= 0 || limit > 10000 {
		limit = 10000
	}

	expr := "id like 'kb%'"

	results, err := m.client.Query(
		ctx,
		m.collection,
		[]string{},
		expr,
		[]string{"id", "title", "content", "category", "tags", "created_at", "created_by", "updated_at", "updated_by"},
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list documents: %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("Milvus Query returned %d documents (limit requested: %d)", results.GetColumn("id").Len(), limit))

	docs := []KnowledgeDocument{}
	idCol := results.GetColumn("id")
	titleCol := results.GetColumn("title")
	contentCol := results.GetColumn("content")
	categoryCol := results.GetColumn("category")
	tagsCol := results.GetColumn("tags")
	createdAtCol := results.GetColumn("created_at")
	createdByCol := results.GetColumn("created_by")
	updatedAtCol := results.GetColumn("updated_at")
	updatedByCol := results.GetColumn("updated_by")

	for i := 0; i < idCol.Len() && i < limit; i++ {
		id, _ := idCol.Get(i)
		title, _ := titleCol.Get(i)
		content, _ := contentCol.Get(i)
		category, _ := categoryCol.Get(i)
		tagsStr, _ := tagsCol.Get(i)
		createdAt, _ := createdAtCol.Get(i)
		createdBy, _ := createdByCol.Get(i)
		updatedAt, _ := updatedAtCol.Get(i)
		updatedBy, _ := updatedByCol.Get(i)

		var tags []string
		if tagsStr != nil {
			json.Unmarshal([]byte(tagsStr.(string)), &tags)
		}

		metadata := map[string]interface{}{}
		if createdAt != nil {
			metadata["created_at"] = createdAt.(string)
		}
		if createdBy != nil {
			metadata["created_by"] = createdBy.(string)
		}
		if updatedAt != nil {
			metadata["updated_at"] = updatedAt.(string)
		}
		if updatedBy != nil {
			metadata["updated_by"] = updatedBy.(string)
		}

		docs = append(docs, KnowledgeDocument{
			ID:       id.(string),
			Title:    title.(string),
			Content:  content.(string),
			Category: category.(string),
			Tags:     tags,
			Metadata: metadata,
		})
	}

	logger.Info(fmt.Sprintf("Listed %d documents", len(docs)))
	return docs, nil
}

func (m *MilvusService) GetDocument(ctx context.Context, docID string) (*KnowledgeDocument, error) {
	expr := fmt.Sprintf("id == '%s'", docID)

	results, err := m.client.Query(
		ctx,
		m.collection,
		[]string{},
		expr,
		[]string{"id", "title", "content", "category", "tags", "created_at", "created_by", "updated_at", "updated_by"},
	)
	if err != nil {
		return nil, err
	}

	if results.GetColumn("id").Len() == 0 {
		return nil, fmt.Errorf("document not found: %s", docID)
	}

	id, _ := results.GetColumn("id").Get(0)
	title, _ := results.GetColumn("title").Get(0)
	content, _ := results.GetColumn("content").Get(0)
	category, _ := results.GetColumn("category").Get(0)
	tagsStr, _ := results.GetColumn("tags").Get(0)
	createdAt, _ := results.GetColumn("created_at").Get(0)
	createdBy, _ := results.GetColumn("created_by").Get(0)
	updatedAt, _ := results.GetColumn("updated_at").Get(0)
	updatedBy, _ := results.GetColumn("updated_by").Get(0)

	var tags []string
	if tagsStr != nil {
		json.Unmarshal([]byte(tagsStr.(string)), &tags)
	}

	metadata := map[string]interface{}{}
	if createdAt != nil {
		metadata["created_at"] = createdAt.(string)
	}
	if createdBy != nil {
		metadata["created_by"] = createdBy.(string)
	}
	if updatedAt != nil {
		metadata["updated_at"] = updatedAt.(string)
	}
	if updatedBy != nil {
		metadata["updated_by"] = updatedBy.(string)
	}

	doc := &KnowledgeDocument{
		ID:       id.(string),
		Title:    title.(string),
		Content:  content.(string),
		Category: category.(string),
		Tags:     tags,
		Metadata: metadata,
	}

	return doc, nil
}

func (m *MilvusService) DeleteDocument(ctx context.Context, docID string) error {
	expr := fmt.Sprintf("id == '%s'", docID)

	err := m.client.Delete(ctx, m.collection, "", expr)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete document: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Document deleted: %s", docID))
	return nil
}

func (m *MilvusService) LoadCollection(ctx context.Context) error {
	err := m.client.LoadCollection(ctx, m.collection, false)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Collection %s loaded", m.collection))
	return nil
}

func (m *MilvusService) Close() error {
	if m.client != nil {
		m.client.Close()
		logger.Info("Milvus client closed")
	}
	return nil
}
