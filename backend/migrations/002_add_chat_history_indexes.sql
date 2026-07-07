-- 为历史上下文查询优化添加索引
-- 创建时间：2026-06-29
-- 目的：优化GetRecentMessages查询性能

-- 为chat_messages表添加复合索引
-- 查询语句：SELECT * FROM chat_messages WHERE session_id = ? ORDER BY created_at DESC LIMIT ?
-- 该索引会显著提升历史消息获取性能

CREATE INDEX IF NOT EXISTS idx_chat_messages_session_created 
ON chat_messages(session_id, created_at DESC);

-- 为chat_sessions表添加索引（如果不存在）
-- 查询语句：SELECT * FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC LIMIT ?
CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_updated 
ON chat_sessions(user_id, updated_at DESC);

-- 注释说明索引用途
COMMENT ON INDEX idx_chat_messages_session_created IS 
'优化历史上下文查询：WHERE session_id = ? ORDER BY created_at DESC';

COMMENT ON INDEX idx_chat_sessions_user_updated IS 
'优化用户会话列表查询：WHERE user_id = ? ORDER BY updated_at DESC';

-- 性能预估：
-- 无索引：O(n) 全表扫描 + 排序
-- 有索引：O(log n) 索引查找 + O(k) 直接返回前k条
-- 预估提升：查询时间从数百毫秒降至数十毫秒

-- 验证索引是否生效：
-- EXPLAIN ANALYZE SELECT * FROM chat_messages 
-- WHERE session_id = 'test-session-id' 
-- ORDER BY created_at DESC LIMIT 20;