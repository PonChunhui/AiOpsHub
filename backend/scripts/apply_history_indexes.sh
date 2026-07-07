#!/bin/bash

# 应用数据库索引优化
# 用于提升历史上下文查询性能

set -e

echo "=== 开始应用历史上下文优化索引 ==="

# 检查数据库连接配置
DB_HOST=${DB_HOST:-"192.168.100.10"}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-"aiops"}
DB_PASSWORD=${DB_PASSWORD:-"aiops123"}
DB_NAME=${DB_NAME:-"aiopsdb"}

echo "数据库连接信息："
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"

# 应用migration文件
MIGRATION_FILE="migrations/002_add_chat_history_indexes.sql"

if [ ! -f "$MIGRATION_FILE" ]; then
    echo "错误：找不到migration文件 $MIGRATION_FILE"
    exit 1
fi

echo ""
echo "正在执行索引创建SQL..."

# 使用psql执行migration
export PGPASSWORD="$DB_PASSWORD"

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"

if [ $? -eq 0 ]; then
    echo "✅ 索引创建成功"
else
    echo "❌ 索引创建失败"
    exit 1
fi

echo ""
echo "=== 验证索引是否创建成功 ==="

# 查询索引信息
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename IN ('chat_messages', 'chat_sessions')
    AND indexname LIKE 'idx_chat_%'
ORDER BY indexname;
"

echo ""
echo "=== 性能测试（可选） ==="
echo "您可以运行以下SQL验证索引效果："
echo ""
echo "EXPLAIN ANALYZE SELECT * FROM chat_messages"
echo "WHERE session_id = 'test-session-id'"
echo "ORDER BY created_at DESC LIMIT 20;"
echo ""

unset PGPASSWORD

echo "=== 完成 ==="