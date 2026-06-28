package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

// EmbeddingService Embedding向量编码服务
// 将文本转换为向量，用于向量检索和语义相似度计算
type EmbeddingService struct {
	Provider       string // Embedding服务提供商（openai, aliyun_bailian等）
	Model          string // Embedding模型名称
	APIKey         string // API密钥
	BaseURL        string // API基础URL
	MaxChunkSize   int    // 最大分块大小（字符数），默认800
	ChunkOverlap   int    // 分块重叠大小（字符数），默认100
	EnableChunking bool   // 是否启用智能分块，默认true
}

// EmbeddingRequest OpenAI兼容格式的Embedding请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse OpenAI兼容格式的Embedding响应
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// AlibabaEmbeddingResponse 阿里云原生格式的Embedding响应
type AlibabaEmbeddingResponse struct {
	Output struct {
		Embeddings []struct {
			TextIndex int       `json:"text_index"`
			Embedding []float32 `json:"embedding"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingService 创建Embedding服务实例
// 参数：provider - 服务提供商, model - 模型名称, apiKey - API密钥, baseURL - API地址
// 返回：Embedding服务实例
func NewEmbeddingService(provider, model, apiKey, baseURL string) *EmbeddingService {
	service := &EmbeddingService{
		Provider:       provider,
		Model:          model,
		APIKey:         apiKey,
		BaseURL:        baseURL,
		MaxChunkSize:   800,  // 默认800字符（低于API限制1024 tokens）
		ChunkOverlap:   100,  // 默认100字符重叠
		EnableChunking: true, // 默认启用分块
	}

	logger.Info(fmt.Sprintf("Embedding service created: provider=%s, model=%s, chunking=%v, chunkSize=%d",
		provider, model, service.EnableChunking, service.MaxChunkSize))
	return service
}

// GetEmbedding 获取文本的Embedding向量
// 支持智能分块处理长文本，避免API截断导致信息丢失
//
// 处理策略：
// 1. 短文本（≤800字符）：直接调用API
// 2. 长文本（>800字符）：
//   - 智能分块：按段落、句子边界分块，保留语义完整性
//   - 分别调用API：为每个分块生成向量
//   - 向量融合：通过平均池化融合多个向量
//
// 参数：ctx - 上下文, text - 文本内容
// 返回：Embedding向量和错误信息
func (e *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// 判断是否需要分块
	if !e.EnableChunking || len(text) <= e.MaxChunkSize {
		// 短文本或未启用分块：直接调用API
		return e.getSingleEmbedding(ctx, text)
	}

	// 长文本：智能分块处理
	logger.Info(fmt.Sprintf("Long text detected, using chunking strategy: length=%d", len(text)))

	// 1. 智能分块
	chunks := e.chunkText(text)
	logger.Info(fmt.Sprintf("Text split into %d chunks", len(chunks)))

	// 2. 为每个分块生成向量
	embeddings := [][]float32{}
	for i, chunk := range chunks {
		logger.Info(fmt.Sprintf("Processing chunk %d: length=%d, preview=%s",
			i+1, len(chunk), truncatePreview(chunk, 50)))

		embedding, err := e.getSingleEmbedding(ctx, chunk)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get embedding for chunk %d: %v", i+1, err))
			// 继续处理其他分块，不中断整个流程
			continue
		}
		embeddings = append(embeddings, embedding)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("failed to get embeddings for all chunks")
	}

	// 3. 向量融合（平均池化）
	fusedEmbedding := e.averagePooling(embeddings)
	logger.Info(fmt.Sprintf("Embeddings fused: chunks=%d, dimension=%d",
		len(embeddings), len(fusedEmbedding)))

	return fusedEmbedding, nil
}

// getSingleEmbedding 获取单个文本块的Embedding（不分块）
// 内部方法，处理API调用
func (e *EmbeddingService) getSingleEmbedding(ctx context.Context, text string) ([]float32, error) {
	switch e.Provider {
	case "openai":
		return e.getOpenAIEmbedding(ctx, text)
	case "aliyun_bailian", "qwen":
		// 阿里云使用OpenAI兼容模式
		return e.getOpenAIEmbedding(ctx, text)
	default:
		return e.getMockEmbedding(text), nil
	}
}

// chunkText 智能分块文本
// 优先按段落边界分块，然后按句子边界，确保语义完整性
//
// 分块策略：
// 1. 优先按段落分割（双换行符）
// 2. 如果段落过长，按句子分割（句号、问号、感叹号）
// 3. 添加重叠区域，避免语义断裂
//
// 参数：text - 待分块的文本
// 返回：分块列表
func (e *EmbeddingService) chunkText(text string) []string {
	chunks := []string{}

	// 1. 先按段落分割（优先保持段落完整性）
	paragraphs := strings.Split(text, "\n\n")

	currentChunk := ""
	for _, paragraph := range paragraphs {
		// 如果当前块加入这个段落不超过限制，直接添加
		if len(currentChunk)+len(paragraph)+2 <= e.MaxChunkSize {
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += paragraph
		} else {
			// 当前块已满，保存并开始新块
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}

			// 如果段落本身过长，需要进一步按句子分割
			if len(paragraph) > e.MaxChunkSize {
				sentenceChunks := e.chunkBySentences(paragraph)
				chunks = append(chunks, sentenceChunks...)
				currentChunk = ""
			} else {
				currentChunk = paragraph
			}
		}
	}

	// 保存最后一个块
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	// 2. 添加重叠区域（提高语义连贯性）
	if e.ChunkOverlap > 0 && len(chunks) > 1 {
		chunks = e.addOverlap(chunks)
	}

	return chunks
}

// chunkBySentences 按句子分割长段落
// 当段落过长时，按句子边界进一步分割
func (e *EmbeddingService) chunkBySentences(text string) []string {
	chunks := []string{}

	// 中文和英文句子分隔符
	separators := []string{"。", "！", "？", ".", "!", "?"}

	// 找到所有句子分隔点
	sentences := []string{}
	lastPos := 0

	for i := 0; i < len(text); i++ {
		for _, sep := range separators {
			if i+len(sep) <= len(text) && text[i:i+len(sep)] == sep {
				sentence := text[lastPos : i+len(sep)]
				sentences = append(sentences, sentence)
				lastPos = i + len(sep)
				break
			}
		}
	}

	// 添加剩余部分
	if lastPos < len(text) {
		sentences = append(sentences, text[lastPos:])
	}

	// 合并句子为块，确保不超过限制
	currentChunk := ""
	for _, sentence := range sentences {
		if len(currentChunk)+len(sentence) <= e.MaxChunkSize {
			currentChunk += sentence
		} else {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			currentChunk = sentence
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// addOverlap 为相邻块添加重叠区域
// 提高语义连贯性，避免边界信息丢失
func (e *EmbeddingService) addOverlap(chunks []string) []string {
	if len(chunks) <= 1 {
		return chunks
	}

	overlappedChunks := []string{}
	for i, chunk := range chunks {
		enhancedChunk := chunk

		// 添加前一块的结尾部分（重叠）
		if i > 0 {
			prevChunk := chunks[i-1]
			overlapStart := len(prevChunk) - e.ChunkOverlap
			if overlapStart < 0 {
				overlapStart = 0
			}
			overlap := prevChunk[overlapStart:]
			enhancedChunk = overlap + enhancedChunk
		}

		// 添加后一块的开头部分（重叠）
		if i < len(chunks)-1 {
			nextChunk := chunks[i+1]
			overlapEnd := e.ChunkOverlap
			if overlapEnd > len(nextChunk) {
				overlapEnd = len(nextChunk)
			}
			overlap := nextChunk[:overlapEnd]
			enhancedChunk = enhancedChunk + overlap
		}

		overlappedChunks = append(overlappedChunks, enhancedChunk)
	}

	return overlappedChunks
}

// averagePooling 平均池化融合多个向量
// 计算多个Embedding向量的平均值，得到融合向量
//
// 参数：embeddings - 多个Embedding向量列表
// 返回：融合后的单个向量
func (e *EmbeddingService) averagePooling(embeddings [][]float32) []float32 {
	if len(embeddings) == 0 {
		return []float32{}
	}

	if len(embeddings) == 1 {
		return embeddings[0]
	}

	// 获取向量维度
	dimension := len(embeddings[0])
	fused := make([]float32, dimension)

	// 计算每个维度的平均值
	for i := 0; i < dimension; i++ {
		sum := float32(0)
		for _, embedding := range embeddings {
			if i < len(embedding) {
				sum += embedding[i]
			}
		}
		fused[i] = sum / float32(len(embeddings))
	}

	return fused
}

// truncatePreview 截取文本预览（用于日志）
func truncatePreview(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// getOpenAIEmbedding 获取OpenAI格式的Embedding（OpenAI和阿里云兼容模式）
// 内部方法，处理HTTP请求和响应解析
func (e *EmbeddingService) getOpenAIEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Model: e.Model,
		Input: []string{text},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", e.BaseURL+"/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("OpenAI embedding request failed: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		logger.Error(fmt.Sprintf("OpenAI embedding failed: status=%d, body=%s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("embedding request failed: status %d", resp.StatusCode)
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, err
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	logger.Info(fmt.Sprintf("OpenAI embedding generated: tokens=%d", embeddingResp.Usage.TotalTokens))
	return embeddingResp.Data[0].Embedding, nil
}

// getMockEmbedding 生成Mock Embedding（用于测试）
// 当没有可用的Embedding服务时，生成虚拟向量
func (e *EmbeddingService) getMockEmbedding(text string) []float32 {
	dimension := 1536
	embedding := make([]float32, dimension)

	for i := range embedding {
		embedding[i] = float32(i%10) / 10.0
	}

	logger.Info(fmt.Sprintf("Mock embedding generated for text (length=%d)", len(text)))
	return embedding
}
