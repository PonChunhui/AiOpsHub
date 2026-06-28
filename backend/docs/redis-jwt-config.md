# Redis JWT Token配置说明

## Redis模式配置

AiOpsHub支持两种Redis部署模式：**单机模式**和**集群模式**。

### 单机模式配置

```yaml
redis:
  cluster_mode: false  # 或不设置此字段
  host: "192.168.100.114"
  port: 6379
  password: "your_password"
  db: 0
```

### 集群模式配置

```yaml
redis:
  cluster_mode: true
  cluster_nodes:
    - "192.168.100.113:6379"
    - "192.168.100.114:6379"
    - "192.168.100.115:6379"
    - "192.168.100.116:6379"
    - "192.168.100.117:6379"
    - "192.168.100.118:6379"
  password: "your_password"
  db: 0
```

**重要说明**：
- 集群模式下，需要配置**所有Redis节点地址**
- Redis Cluster会自动处理key的slot分配和请求重定向
- `cluster_mode: true` 时，`host`和`port`字段会被忽略，只使用`cluster_nodes`

## JWT Token配置

```yaml
jwt:
  secret: "aiops-secret-key-change-in-production"  # JWT签名密钥
  token_expire: 30m  # Token有效期，支持时间单位：s(秒), m(分钟), h(小时)
```

**配置示例**：
- `30m` - 30分钟
- `1h` - 1小时
- `24h` - 24小时
- `3600s` - 3600秒

## Token存储机制

### Redis存储结构

Token存储在Redis中，key格式为：`token:{jwt_token_string}`

存储的TokenInfo结构（JSON格式）：

```json
{
  "user_id": "cbc3af2d-5bde-4608-a62a-9601f9973264",
  "username": "admin",
  "role": "user",
  "token_type": "access",
  "created_at": "2026-06-24T23:17:33Z",
  "source": "login"
}
```

**字段说明**：
- `user_id`: 用户唯一ID
- `username`: 用户名
- `role`: 用户角色
- `token_type`: Token类型（目前固定为"access"）
- `created_at`: Token创建时间
- `source`: Token来源（登录、刷新等）

### 验证流程

1. **JWT签名验证**：验证token的签名是否有效
2. **Redis存在性验证**：检查token是否存在于Redis中
3. **过期检查**：Redis中token会自动过期（根据`jwt.token_expire`配置）

### 双重验证机制

- **JWT验证**：快速验证token签名和claims
- **Redis验证**：确保token未被主动注销（logout）

即使JWT本身有效，如果Redis中不存在该token，验证也会失败。

## API接口

### 登录
```
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "admin123"
}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "...",
    "username": "admin",
    "role": "user",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 注销
```
POST /api/v1/auth/logout
Authorization: Bearer {token}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "logout successful"
  }
}
```

### 使用Token访问API
```
GET /api/v1/workflows
Authorization: Bearer {token}
```

## 错误处理

### Redis连接失败

系统会自动降级处理：
- 如果Redis连接失败，系统会输出Warning日志
- Token验证会降级为纯JWT验证（不检查Redis存在性）
- 系统仍可正常运行，但缺少logout功能支持

### Token验证失败

```json
{
  "error": "missing authorization token"
}
```
- 缺少Authorization header

```json
{
  "error": "invalid or expired token: token is malformed: ..."
}
```
- Token格式错误

```json
{
  "error": "invalid or expired token: token not found in Redis or expired"
}
```
- Token已过期或已被注销

## 安全建议

1. **生产环境必须修改JWT密钥**
   ```yaml
   jwt:
     secret: "your-production-secret-key-at-least-32-characters"
   ```

2. **合理设置Token过期时间**
   - 开发环境：可以设置较长（如1小时）
   - 生产环境：建议30分钟，配合刷新机制

3. **Redis安全配置**
   - 设置强密码
   - 网络隔离
   - 定期备份

4. **HTTPS部署**
   - 生产环境必须使用HTTPS传输token
   - 防止token被中间人攻击窃取

## 测试验证

### 测试登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 测试Token访问
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

curl http://localhost:8080/api/v1/workflows \
  -H "Authorization: Bearer $TOKEN"
```

### 测试注销
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

## 实现文件

- **配置文件**: `backend/config/config.yaml`
- **Redis客户端**: `backend/pkg/redis/redis.go`
- **JWT工具**: `backend/pkg/jwt/jwt.go`
- **认证中间件**: `backend/internal/middleware/middleware.go`
- **登录/注销Handler**: `backend/internal/handler/handler.go`