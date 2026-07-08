# SSH白名单配置示例指南

## 问题回顾

**原始错误：**
```
命令 'ls -lh /var/log/' 不在白名单中
允许的命令: [ls top free df ps netstat cat /var/log/*]
```

**根本原因：**
1. 配置只有`ls`（精确命令），但用户执行`ls -lh /var/log/`（带参数）
2. 旧的白名单验证不支持命令前缀匹配

---

## 已实施的修复方案

### 修复内容

**文件：** `backend/internal/agent/eino_tools/ssh_tool.go`

**核心改进：**
1. **命令前缀匹配** - `ls`可以匹配`ls -lh`, `ls -la`, `ls /var/log`等
2. **参数通配符匹配** - `cat /var/log/*`匹配`cat /var/log/messages`等
3. **多层次匹配策略** - 精确匹配、前缀匹配、通配符匹配

---

## 白名单配置规则详解

### 规则1：全局通配符

**配置：**
```json
"allowed_commands": ["*"]
```

**匹配：** 任何命令都可以执行

**使用场景：** 完全信任的Agent（谨慎使用）

---

### 规则2：精确命令匹配

**配置：**
```json
"allowed_commands": ["ls", "top", "free"]
```

**匹配：**
- `ls` → ✅ 匹配所有ls命令（ls, ls -l, ls /tmp, ls -lh /var/log等）
- `top` → ✅ 匹配所有top命令（top, top -n 1等）
- `free` → ✅ 匹配所有free命令（free, free -h等）

**匹配逻辑：** 只要命令的第一部分匹配，允许执行

**示例：**
| 配置 | 用户执行 | 是否匹配 |
|------|---------|---------|
| `"ls"` | `ls` | ✅ |
| `"ls"` | `ls -l` | ✅ |
| `"ls"` | `ls -lh /var/log` | ✅ |
| `"ls"` | `lsabc` | ❌（不是ls命令） |

---

### 规则3：命令+路径通配符

**配置：**
```json
"allowed_commands": ["cat /var/log/*"]
```

**匹配逻辑：**
- 基础命令：`cat`
- 路径模式：`/var/log/*`
- 只要命令是`cat`且路径以`/var/log/`开头，允许执行

**示例：**
| 配置 | 用户执行 | 是否匹配 | 说明 |
|------|---------|---------|------|
| `cat /var/log/*` | `cat /var/log/messages` | ✅ | 精确匹配路径 |
| `cat /var/log/*` | `cat /var/log/syslog` | ✅ | 精确匹配路径 |
| `cat /var/log/*` | `cat /var/log/nginx/access.log` | ✅ | 路径前缀匹配 |
| `cat /var/log/*` | `cat /etc/passwd` | ❌ | 路径不匹配 |
| `cat /var/log/*` | `ls /var/log/messages` | ❌ | 命令不匹配 |

---

### 规则4：中间通配符

**配置：**
```json
"allowed_commands": ["tail /var/log/*/error.log"]
```

**匹配：**
- `tail /var/log/nginx/error.log` → ✅
- `tail /var/log/mysql/error.log` → ✅
- `tail /var/log/app/error.log` → ✅

---

### 规则5：多重通配符

**配置：**
```json
"allowed_commands": ["cat /tmp/*/*.log"]
```

**匹配：**
- `cat /tmp/app/error.log` → ✅
- `cat /tmp/nginx/access.log` → ✅
- `cat /var/log/messages` → ❌（路径不匹配）

---

## 推荐配置示例

### 1. 系统巡检Agent（宽松）

**推荐配置：**
```json
{
  "allowed_commands": [
    "ls",        // 允许所有ls命令
    "top",       // 允许所有top命令  
    "free",      // 允许所有free命令
    "df",        // 允许所有df命令
    "ps",        // 允许所有ps命令
    "netstat",   // 允许所有netstat命令
    "cat /var/log/*",  // 允许查看/var/log下的所有文件
    "tail /var/log/*", // 允许tail /var/log下的所有文件
    "grep"       // 允许所有grep命令
  ],
  "allowed_hosts": ["*"],
  "timeout": 30
}
```

**适用场景：** 运维巡检，需要广泛访问权限

---

### 2. 告警分析Agent（中等）

**推荐配置：**
```json
{
  "allowed_commands": [
    "cat /var/log/messages",  // 只允许查看messages文件
    "cat /var/log/syslog",     // 只允许查看syslog文件
    "tail /var/log/*",         // 允许tail所有日志文件
    "grep",                    // 允许grep命令
    "ps",                      // 允许查看进程
    "netstat"                  // 允许查看网络状态
  ],
  "allowed_hosts": [
    "192.168.100.10",
    "192.168.100.20"
  ],
  "timeout": 60
}
```

**适用场景：** 告警处理，限定文件和主机范围

---

### 3. 日志查询Agent（严格）

**推荐配置：**
```json
{
  "allowed_commands": [
    "cat /var/log/app/*.log",      // 只允许查看app目录下的日志
    "tail /var/log/app/*.log",     // 只允许tail app日志
    "grep /var/log/app/*.log"      // 只允许grep app日志
  ],
  "allowed_hosts": [
    "app-server-01",
    "app-server-02"
  ],
  "timeout": 30
}
```

**适用场景：** 应用日志分析，严格限定范围

---

### 4. 服务器命令Agent（受限）

**推荐配置：**
```json
{
  "allowed_commands": [
    "ls /tmp",           // 只允许ls /tmp目录
    "cat /tmp/*.txt",    // 只允许查看/tmp下的txt文件
    "rm /tmp/*.tmp"      // 只允许删除/tmp下的tmp文件
  ],
  "allowed_hosts": [
    "192.168.100.186"
  ],
  "timeout": 15
}
```

**适用场景：** 临时文件管理，严格限制操作范围

---

## 匹配逻辑详解

### 匹配流程

```
用户命令：ls -lh /var/log/
↓
步骤1：分离命令和参数
  命令：ls
  参数：-lh /var/log/
↓
步骤2：遍历白名单
  配置：ls（只有命令）
↓
步骤3：命令匹配
  配置命令：ls
  用户命令：ls
  → 匹配成功！
↓
步骤4：参数检查
  配置只有命令，无参数限制
  → 允许所有参数
↓
结果：✅ 命令允许执行
```

### 另一个示例

```
用户命令：cat /var/log/messages
↓
步骤1：分离
  命令：cat
  参数：/var/log/messages
↓
步骤2：遍历白名单
  配置：cat /var/log/*
↓
步骤3：命令匹配
  配置命令：cat
  用户命令：cat
  → 匹配成功
↓
步骤4：参数匹配
  配置参数：/var/log/*
  用户参数：/var/log/messages
  filepath.Match("/var/log/*", "/var/log/messages")
  → 匹配成功
↓
结果：✅ 命令允许执行
```

---

## 配置SQL示例

### 插入Agent工具绑定

```sql
INSERT INTO agent_tools (
  id, 
  agent_id, 
  tool_id, 
  config_override, 
  enabled, 
  priority, 
  created_at, 
  updated_at
) VALUES (
  'binding-sys-inspection-ssh',
  'preset-system-inspection',
  'tool-ssh-exec',
  '{"allowed_commands": ["ls", "top", "free", "df", "ps", "netstat", "cat /var/log/*", "tail /var/log/*"], "allowed_hosts": ["*"], "timeout": 30}',
  true,
  10,
  NOW(),
  NOW()
);
```

### 更新现有绑定

```sql
UPDATE agent_tools 
SET config_override = '{"allowed_commands": ["ls", "top", "free", "df", "ps", "cat /var/log/*"], "allowed_hosts": ["192.168.100.*"], "timeout": 60}'
WHERE agent_id = 'preset-alert-handler' 
  AND tool_id LIKE '%ssh%';
```

---

## 测试验证

### 测试用例

| Agent | 配置命令 | 测试命令 | 预期结果 |
|-------|---------|---------|---------|
| 系统巡检 | `ls` | `ls -lh /var/log/` | ✅ 允许 |
| 系统巡检 | `ls` | `ls -la /tmp` | ✅ 允许 |
| 系统巡检 | `cat /var/log/*` | `cat /var/log/messages` | ✅ 允许 |
| 系统巡检 | `cat /var/log/*` | `cat /etc/passwd` | ❌ 拒绝 |
| 告警分析 | `cat /var/log/messages` | `cat /var/log/messages` | ✅ 允许 |
| 告警分析 | `cat /var/log/messages` | `cat /var/log/syslog` | ❌ 拒绝 |

### 测试SQL查询

```sql
-- 查看当前配置
SELECT 
  a.name as agent_name,
  t.name as tool_name,
  at.config_override
FROM agent_tools at
JOIN agents a ON at.agent_id = a.id
JOIN tools t ON at.tool_id = t.id
WHERE t.name = 'ssh_exec'
ORDER BY a.name;
```

---

## 常见问题排查

### 问题1：命令仍然被拒绝

**排查步骤：**

1. **检查配置格式**
   ```sql
   SELECT config_override FROM agent_tools 
   WHERE agent_id = 'your-agent-id' AND tool_id LIKE '%ssh%';
   ```
   确认config_override是正确的JSON格式

2. **检查命令拼写**
   - 配置中的命令名称是否正确
   - 是否有多余空格

3. **查看执行日志**
   ```
   工具 ssh_exec 执行成功，耗时Xms
   或
   命令 'xxx' 不在白名单中，允许的命令: [...]
   ```

### 问题2：配置未生效

**原因：** ToolRegistry缓存了旧的工具实例

**解决：** 清空Agent缓存
```bash
curl -X POST http://localhost:8080/api/v1/admin/agent/cache/clear \
  -H "Authorization: Bearer <token>"
```

或重启服务

---

## 安全建议

### ✅ 推荐

1. **使用精确配置**：明确指定允许的命令和路径
2. **限制主机范围**：使用`allowed_hosts`限定目标服务器
3. **设置超时时间**：避免长时间运行的命令
4. **分级授权**：不同Agent给予不同权限级别

### ⚠️ 谨慎使用

1. **全局通配符 `"*"`**：允许所有命令，风险最高
2. **全局主机 `"*"`**：允许访问所有服务器
3. **危险命令**：`rm`, `chmod`, `chown`, `shutdown`等

### ❌ 不推荐

1. **允许rm命令**：可能误删重要文件
2. **允许chmod/chown**：可能改变文件权限
3. **允许sudo**：可能提权执行危险命令

---

## 配置模板

### 最小权限模板

```json
{
  "allowed_commands": ["ls", "cat /var/log/*.log"],
  "allowed_hosts": ["specific-host-id"],
  "timeout": 15
}
```

### 中等权限模板

```json
{
  "allowed_commands": ["ls", "top", "free", "cat /var/log/*", "tail /var/log/*"],
  "allowed_hosts": ["192.168.100.*"],
  "timeout": 30
}
```

### 最大权限模板（慎用）

```json
{
  "allowed_commands": ["*"],
  "allowed_hosts": ["*"],
  "timeout": 60
}
```

---

## 总结

### 关键改进

✅ **命令前缀匹配** - `ls`匹配所有ls命令及其参数
✅ **路径通配符** - `/var/log/*`匹配/var/log下的所有文件
✅ **灵活配置** - 支持多种匹配策略组合

### 配置建议

1. **精确配置优于宽泛配置**
2. **路径限制优于命令限制**
3. **主机限制优于全局允许**
4. **超时设置避免长时间占用**

---

**文档位置：** `docs/ssh-whitelist-configuration-guide.md`
**更新时间：** 2026-07-07
**状态：** ✅ 修复完成，建议按推荐配置更新agent_tools表