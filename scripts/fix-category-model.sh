#!/bin/bash
# 批量修复知识库模型修改的脚本

echo "开始批量修复Category -> DocType+Component..."

# 修改rag_service.go中的示例文档
cd backend

# 1. 修改示例知识库文档
sed -i.bak '
230,266 {
  s/Category: "troubleshooting"/DocType: "sop",\n\t\tComponent: "general"/g
  s/Category: "optimization"/DocType: "sop",\n\t\tComponent: "mysql"/g  
  s/Category: "kubernetes"/DocType: "sop",\n\t\tComponent: "k8s"/g
}
' internal/service/rag_service.go

# 2. 修改AddDocumentToKnowledgeBase方法中的字段映射
sed -i.bak '
446 {
  s/Category:  doc.Category/DocType:   doc.DocType,\n\t\tComponent: doc.Component/g
}
' internal/service/rag_service.go

# 3. 修改UpdateDocument方法参数签名和实现
sed -i.bak '
529 {
  s/category string/docType string, component string/g
}
541 {
  s/Category: category/DocType: docType,\n\t\tComponent: component/g  
}
557 {
  s/pgDoc.Category = category/pgDoc.DocType = docType\n\t\t\t\tpgDoc.Component = component/g
}
' internal/service/rag_service.go

# 4. 修改ListDocuments方法参数签名和实现
sed -i.bak '
607 {
  s/category string/docType string, component string/g
}
608 {
  s/category=%s/docType=%s, component=%s/g
}
611 {
  s/category, search/docType, component, search/g
}
' internal/service/rag_service.go

# 5. 修改matchQuery方法中的Category引用
sed -i.bak '
313 {
  s/doc.Category/doc.DocType + "/" + doc.Component/g
}
344 {
  s/doc.Category/doc.DocType + "/" + doc.Component/g  
}
' internal/service/rag_service.go

echo "rag_service.go修复完成"

# 修改rag_repo.go
sed -i.bak '
30 {
  s/category string/docType string, component string/g
}
35,36 {
  s/category/docType/g
  s/"category = ?"/"(doc_type = ? OR doc_type IS NULL)" \&\& (component = ? OR component IS NULL)"/g
}
' internal/repository/rag_repo.go

echo "rag_repo.go修复完成"

# 检查编译
echo "检查编译..."
go build ./cmd/api-server 2>&1 | grep -E "Category|doc_type|component" | head -5

echo "修复完成！如果还有错误，请手动修复。"
echo "备份文件位置：backend/*.bak"