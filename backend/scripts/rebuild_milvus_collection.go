package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

func main() {
	// 加载配置
	config.Init()

	// 初始化日志
	logger.Init()

	// 初始化数据库
	database.Init()

	// 连接Milvus
	milvusHost := viper.GetString("milvus.host")
	milvusPort := viper.GetString("milvus.port")

	ctx := context.Background()

	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: fmt.Sprintf("%s:%s", milvusHost, milvusPort),
	})
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}

	collectionName := "knowledge_documents"

	// 检查并删除旧collection
	has, err := milvusClient.HasCollection(ctx, collectionName)
	if err != nil {
		log.Printf("检查collection失败: %v", err)
	} else if has {
		fmt.Println("⚠️  删除旧的collection...")
		err = milvusClient.DropCollection(ctx, collectionName)
		if err != nil {
			log.Printf("删除collection失败: %v", err)
		} else {
			fmt.Println("✅ 旧collection已删除")
		}
	}

	// 创建新collection（包含doc_type和component）
	fmt.Println("🔨 创建新的collection（包含doc_type和component字段）...")

	schema := &entity.Schema{
		CollectionName: collectionName,
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
				Name:       "doc_type",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50"},
			},
			{
				Name:       "component",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50"},
			},
			{
				Name:       "tags",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "500"},
			},
			{
				Name:       "embedding",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{"dim": "1536"},
			},
			{
				Name:       "created_at",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50"},
			},
			{
				Name:       "created_by",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "updated_at",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50"},
			},
			{
				Name:       "updated_by",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
		},
	}

	err = milvusClient.CreateCollection(ctx, schema, 1)
	if err != nil {
		log.Fatalf("创建collection失败: %v", err)
	}
	fmt.Println("✅ 新collection已创建")

	// 创建索引
	fmt.Println("📊 创建向量索引...")
	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 128)
	if err != nil {
		log.Printf("创建索引对象失败: %v", err)
	} else {
		err = milvusClient.CreateIndex(ctx, collectionName, "embedding", idx, false)
		if err != nil {
			log.Printf("创建索引失败: %v", err)
		} else {
			fmt.Println("✅ 索引已创建")
		}
	}

	// 加载collection
	fmt.Println("📦 加载collection到内存...")
	err = milvusClient.LoadCollection(ctx, collectionName, false)
	if err != nil {
		log.Printf("加载collection失败: %v", err)
	} else {
		fmt.Println("✅ Collection已加载")
	}

	// 从PostgreSQL查询数据
	fmt.Println("🔄 从PostgreSQL读取文档...")
	var docs []model.RAGDocument
	result := database.DB.Find(&docs)
	if result.Error != nil {
		log.Fatalf("查询失败: %v", result.Error)
	}

	fmt.Printf("找到%d条文档\n", len(docs))

	if len(docs) == 0 {
		fmt.Println("⚠️  PostgreSQL中没有文档数据")
		fmt.Println("✅ Milvus collection已重建完成")
		return
	}

	// 批量插入（暂时不生成向量，只插入元数据）
	fmt.Println("💾 插入文档到Milvus...")

	// 为每条文档生成临时向量（全0向量，后续可重新生成）
	tempEmbedding := make([]float32, 1536)

	successCount := 0
	for i, doc := range docs {
		tagsJSON, _ := json.Marshal([]string{})
		if doc.Tags != "" {
			json.Unmarshal([]byte(doc.Tags), &tagsJSON)
		}

		columns := []entity.Column{
			entity.NewColumnVarChar("id", []string{doc.ID}),
			entity.NewColumnVarChar("title", []string{doc.Title}),
			entity.NewColumnVarChar("content", []string{doc.Content}),
			entity.NewColumnVarChar("doc_type", []string{doc.DocType}),
			entity.NewColumnVarChar("component", []string{doc.Component}),
			entity.NewColumnVarChar("tags", []string{doc.Tags}),
			entity.NewColumnFloatVector("embedding", 1536, [][]float32{tempEmbedding}),
			entity.NewColumnVarChar("created_at", []string{doc.CreatedAt.Format(time.RFC3339)}),
			entity.NewColumnVarChar("created_by", []string{doc.CreatedBy}),
			entity.NewColumnVarChar("updated_at", []string{doc.UpdatedAt.Format(time.RFC3339)}),
			entity.NewColumnVarChar("updated_by", []string{doc.UpdatedBy}),
		}

		_, err := milvusClient.Insert(ctx, collectionName, "", columns...)
		if err != nil {
			logger.Error(fmt.Sprintf("[%d] 插入失败: %v", i+1, err))
			continue
		}

		successCount++
		if (i+1)%10 == 0 {
			fmt.Printf("进度: %d/%d\n", i+1, len(docs))
		}
	}

	fmt.Printf("\n✅ 成功插入%d条文档\n", successCount)
	fmt.Println("🎉 Milvus重建完成！")
	fmt.Println("\n⚠️  注意：向量embedding暂时为空向量，需要后续重新生成")
}
