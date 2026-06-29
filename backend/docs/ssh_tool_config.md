# SSH 工具配置指南

## 配置方式

SSH 工具支持两种认证方式：密码认证和密钥认证。

### 1. 密码认证配置

```bash
TOKEN="<your_token>"
curl -X PUT "http://127.0.0.1:8080/api/v1/agents/preset-server-command/tools/tool-ssh-exec/config" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config_override": {
      "username": "root",
      "password": "your_password",
      "port": 22,
      "timeout": 30,
      "allowed_hosts": ["192.168.100.10", "192.168.1.100"],
      "allowed_commands": ["ls", "cat", "df -h", "top", "ps", "free", "netstat"]
    }
  }'
```

### 2. 密钥认证配置

```bash
TOKEN="<your_token>"
curl -X PUT "http://127.0.0.1:8080/api/v1/agents/preset-server-command/tools/tool-ssh-exec/config" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config_override": {
      "username": "root",
      "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
      "port": 22,
      "timeout": 30,
      "allowed_hosts": ["192.168.100.10"],
      "allowed_commands": ["ls", "df -h", "cat"]
    }
  }'
```

### 3. 读取私钥文件

```bash
# 从文件读取私钥
PRIVATE_KEY=$(cat ~/.ssh/id_rsa)

curl -X PUT "http://127.0.0.1:8080/api/v1/agents/preset-server-command/tools/tool-ssh-exec/config" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"config_override\": {
      \"username\": \"root\",
      \"private_key\": \"$PRIVATE_KEY\",
      \"allowed_hosts\": [\"192.168.100.10\"],
      \"allowed_commands\": [\"ls\", \"df -h\"]
    }
  }"
```

## 配置字段说明

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `username` | string | 否 | `root` | SSH登录用户名 |
| `password` | string | 条件 | 无 | SSH密码（密码认证必填） |
| `private_key` | string | 条件 | 无 | SSH私钥（密钥认证必填） |
| `port` | int | 否 | `22` | SSH端口 |
| `timeout` | int | 否 | `30` | 超时时间（秒） |
| `allowed_hosts` | []string | 是 | 无 | 允许连接的主机列表 |
| `allowed_commands` | []string | 是 | 无 | 允许执行的命令白名单 |

## 安全建议

1. **密钥认证优先**：推荐使用密钥认证，更安全
2. **白名单限制**：
   - 主机白名单：只允许特定服务器IP
   - 命令白名单：只允许安全的巡检命令
   - 避免通配符 `*`
3. **命令选择**：
   - 推荐：`ls`, `df -h`, `free`, `top`, `ps`, `netstat`
   - 避免：`rm`, `chmod`, `chown`, `vi`, `nano`等修改性命令
4. **超时设置**：建议设置30-60秒超时，防止长时间等待

## 测试示例

```bash
# 创建测试会话
SESSION_ID=$(curl -s -X POST http://127.0.0.1:8080/api/v1/chat/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"SSH测试","model":"qwen3.7-max"}' | jq -r '.data.id')

# 发送测试消息
curl -X POST http://127.0.0.1:8080/api/v1/chat/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"$SESSION_ID\",\"content\":\"检查服务器192.168.100.10的磁盘空间，执行df -h命令\"}"
```

## 真实实现特性

✅ **已实现**：
- Go SSH客户端（golang.org/x/crypto/ssh）
- 密码和密钥双认证支持
- 主机和命令白名单验证
- 超时控制
- 错误处理和日志记录
- stdout/stderr分离输出

🔒 **安全机制**：
- 白名单强制验证（无法绕过）
- 命令注入检测（检测`; | & $`等危险字符）
- 路径穿越检测（检测`..`）
- HostKeyCallback配置（生产环境需自定义）

⚠️ **生产注意事项**：
- 当前使用 `ssh.InsecureIgnoreHostKey()`，生产环境需配置HostKey验证
- 密码/私钥存储在数据库，建议加密存储
- 建议添加SSH审计日志（记录所有执行命令）