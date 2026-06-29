package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

func main() {
	config.Init()
	logger.Init()
	
	host := viper.GetString("milvus.host")
	port := viper.GetString("milvus.port")
	
	c, err := client.NewClient(context.Background(), client.Config{
		Address: fmt.Sprintf("%s:%s", host, port),
	})
	if err != nil {
		log.Fatal(err)
	}
	
	ctx := context.Background()
	
	desc, err := c.DescribeCollection(ctx, "knowledge_documents")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Collection: %s\n", desc.Name)
	fmt.Printf("Fields (%d):\n", len(desc.Schema.Fields))
	for _, field := range desc.Schema.Fields {
		fmt.Printf("  - %s: %v\n", field.Name, field.DataType)
	}
	
	// 检查是否还有category字段
	hasCategory := false
	for _, field := range desc.Schema.Fields {
		if field.Name == "category" {
			hasCategory = true
			break
		}
	}
	
	if hasCategory {
		fmt.Println("\n❌ ERROR: Collection still has 'category' field!")
	} else {
		fmt.Println("\n✅ Collection has correct fields (no 'category')")
	}
	
	c.Close()
}
