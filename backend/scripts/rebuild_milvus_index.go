package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/spf13/viper"
)

func main() {
	logger.Init()

	log.Println("=== Milvus索引重建脚本 ===")
	log.Println("此脚本将删除现有collection并重新创建（使用余弦相似度）")
	log.Println("警告：所有向量数据将丢失，需要重新从PostgreSQL同步")
	log.Println()

	viper.SetConfigFile("../configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	milvusHost := viper.GetString("milvus.host")
	milvusPort := viper.GetString("milvus.port")

	if milvusHost == "" || milvusPort == "" {
		log.Fatal("Milvus配置缺失，请检查config.yaml中的milvus.host和milvus.port")
	}

	log.Printf("Milvus地址: %s:%s\n", milvusHost, milvusPort)

	milvusSvc, err := service.NewMilvusService(milvusHost, milvusPort, "default")
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}
	defer milvusSvc.Close()

	ctx := context.Background()
	collectionName := "knowledge_documents"

	log.Println("步骤1：检查collection是否存在...")
	has, err := milvusSvc.HasCollection(ctx, collectionName)
	if err != nil {
		log.Fatalf("Failed to check collection: %v", err)
	}

	if has {
		log.Println("步骤2：删除现有collection...")
		log.Println("警告：所有数据将被删除！")

		fmt.Print("确认删除collection并重建？(输入 YES 继续): ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			log.Println("用户取消操作，退出脚本")
			os.Exit(0)
		}

		err := milvusSvc.DropCollection(ctx, collectionName)
		if err != nil {
			log.Fatalf("Failed to drop collection: %v", err)
		}
		log.Println("Collection已删除")
	} else {
		log.Println("Collection不存在，将直接创建")
	}

	log.Println("步骤3：创建新collection（使用余弦相似度）...")
	err = milvusSvc.CreateCollection(ctx)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	log.Println("Collection创建成功（余弦相似度索引）")

	log.Println("步骤4：加载collection到内存...")
	err = milvusSvc.LoadCollection(ctx)
	if err != nil {
		log.Fatalf("Failed to load collection: %v", err)
	}
	log.Println("Collection已加载")

	log.Println()
	log.Println("=== 索引重建完成 ===")
	log.Println("下一步：运行 sync_pg_to_milvus.go 从PostgreSQL重新导入数据")
	log.Println("命令: go run sync_pg_to_milvus.go")
}
