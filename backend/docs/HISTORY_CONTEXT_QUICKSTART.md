# 快速开始：历史上下文功能

## 1分钟快速启用

### 前提条件
- ✅ 已编译后端服务（`go build ./cmd/api-server`）
- ✅ 数据库连接正常
- ✅ 配置文件存在（`configs/config.yaml`）

### 启用步骤

**步骤1: 检查配置**（默认已启用）
```bash
cat configs/config.yaml | grep -A 4 "chat:"
```

**输出**:
```yaml
chat:
  enable_history: true        # 已启用
  max_history_messages: 20    # 最大20条历史
  max_history_tokens: 4000    # 限制4000 tokens
```

**步骤2: 应用数据库索引**（可选，推荐）
```bash
cd backend
./scripts/apply_history_indexes.sh
```

**步骤3: 启动服务**
```bash
./api-server
```

**步骤4: 查看日志确认**
```bash
tail -f backend-new.log | grep "历史上下文"
```

**期望输出**:
```
[历史上下文] 会话xxx: 包含2条历史消息，构建后prompt长度150字符
```

### 测试验证

**快速测试**:
```bash
# 1. 获取token（登录或使用现有token）
export TEST_TOKEN="your-jwt-token"

# 2. 运行集成测试
cd backend
./scripts/test_history_context_integration.sh

# 3. 观察AI是否能记住用户信息
# 第一条消息: "我是张三"
# 第二条消息: "我叫什么名字？"
# 期望回复: "你叫张三"
```

### 配置调整

**如需调整历史消息数量**:
```yaml
chat:
  max_history_messages: 10    # 改为10条
  max_history_tokens: 2000    # 改为2000 tokens
```

**如需临时禁用**:
```yaml
chat:
  enable_history: false       # 禁用历史功能
```

### 问题排查

**问题: AI不记得之前说的内容**

检查:
```bash
# 1. 配置是否启用
grep "enable_history" configs/config.yaml

# 2. 日志是否有错误
tail -100 backend-new.log | grep "历史上下文.*失败"

# 3. 数据库是否有历史消息
psql -c "SELECT COUNT(*) FROM chat_messages WHERE session_id='your-session-id';"
```

**问题: 响应变慢**

检查:
```bash
# 1. 索引是否创建
psql -c "\di chat_messages" | grep idx_chat

# 2. 如果没有，执行索引创建
./scripts/apply_history_indexes.sh
```

## 就这么简单！

历史上下文功能默认已启用，无需额外配置即可使用。

---

**查看完整文档**: `docs/history_context_implementation.md`  
**查看错误处理**: `docs/history_context_error_handling.md`  
**运行完整测试**: `scripts/test_history_context_integration.sh`