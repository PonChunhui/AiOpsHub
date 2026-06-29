-- 知识库表结构迁移脚本
-- 将category字段改为doc_type和component两个字段

-- 1. 添加新字段
ALTER TABLE rag_documents ADD COLUMN IF NOT EXISTS doc_type VARCHAR(50);
ALTER TABLE rag_documents ADD COLUMN IF NOT EXISTS component VARCHAR(50);

-- 2. 从category字段迁移数据到新字段
-- 根据现有的category值，推断doc_type和component
UPDATE rag_documents 
SET 
  doc_type = CASE 
    WHEN category LIKE '%排查%' OR category LIKE '%troubleshooting%' THEN 'sop'
    WHEN category LIKE '%优化%' OR category LIKE '%optimization%' THEN 'sop'
    WHEN category LIKE '%配置%' THEN 'sop'
    WHEN category LIKE '%faq%' OR category LIKE '%问答%' THEN 'faq'
    WHEN category LIKE '%告警%' OR category LIKE '%alert%' THEN 'alert'
    ELSE 'sop'
  END,
  component = CASE 
    WHEN category LIKE '%mysql%' OR category LIKE '%数据库%' OR category LIKE '%database%' THEN 'mysql'
    WHEN category LIKE '%k8s%' OR category LIKE '%kubernetes%' THEN 'k8s'
    WHEN category LIKE '%redis%' OR category LIKE '%缓存%' THEN 'redis'
    WHEN category LIKE '%docker%' THEN 'docker'
    WHEN category LIKE '%nginx%' THEN 'nginx'
    WHEN category LIKE '%java%' OR category LIKE '%go%' THEN 'application'
    ELSE 'general'
  END
WHERE category IS NOT NULL;

-- 3. 为新字段添加索引
CREATE INDEX IF NOT EXISTS idx_rag_documents_doc_type ON rag_documents(doc_type);
CREATE INDEX IF NOT EXISTS idx_rag_documents_component ON rag_documents(component);

-- 4. 删除旧的category字段（可选，建议先保留一段时间）
-- ALTER TABLE rag_documents DROP COLUMN IF EXISTS category;

-- 5. 更新注释
COMMENT ON COLUMN rag_documents.doc_type IS '文档类型：sop / faq / alert';
COMMENT ON COLUMN rag_documents.component IS '组件名：mysql / k8s / redis';

-- 注意事项：
-- 1. 执行前请备份数据库
-- 2. 建议先在第2步测试数据迁移逻辑是否正确
-- 3. 第4步删除category字段建议在确认新字段工作正常后再执行