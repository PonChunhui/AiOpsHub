# 历史上下文功能错误处理和降级策略

## 已实现的错误处理

### 1. 配置层面的错误处理
- **配置缺失**: 如果配置文件中没有chat配置项，会使用默认值
  - `enable_history: true`
  - `max_history_messages: 20`
  - `max_history_tokens: 4000`
  - `max_total_tokens: 8000`

### 2. 数据库查询层面的错误处理
- **获取历史消息失败**: 降级为仅使用当前prompt
  ```go
  allMessages, err := s.repo.GetRecentMessages(sessionID, s.chatConfig.MaxHistoryMessages)
  if err != nil {
      logger.Error(fmt.Sprintf("[历史上下文] 获取历史消息失败: %v，使用降级方案", err))
      return currentPrompt, nil  // 降级为无历史模式
  }
  ```

### 3. 功能开关层面的错误处理
- **历史功能被禁用**: 直接返回当前prompt
  ```go
  if !s.chatConfig.EnableHistory {
      logger.Info(fmt.Sprintf("[历史上下文] 会话%s: 历史上下文功能已禁用", sessionID))
      return currentPrompt, nil
  }
  ```

### 4. 空数据层面的错误处理
- **无历史消息**: 直接返回当前prompt
  ```go
  if len(allMessages) == 0 {
      logger.Info(fmt.Sprintf("[历史上下文] 会话%s无历史消息", sessionID))
      return currentPrompt, nil
  }
  ```

### 5. Token限制层面的错误处理
- **历史消息超长**: 智能截断，保留最新消息
  ```go
  if totalTokens + estimatedTokens > maxTokens {
      logger.Info(fmt.Sprintf("[历史截断] 达到token限制(%d)，保留%d条消息", maxTokens, len(result)))
      break
  }
  ```

## 降级策略优先级

1. **最高优先级**: 功能开关（`enable_history`）
2. **高优先级**: 数据库查询错误降级
3. **中优先级**: 空数据降级
4. **低优先级**: Token截断（不完全降级，只是减少历史）

## 性能降级指标

当以下情况发生时，系统会自动降级：

| 场景 | 降级行为 | 影响 |
|------|---------|------|
| 配置缺失 | 使用默认值 | 无影响，功能正常 |
| 数据库查询失败 | 无历史模式 | AI无法理解对话上下文 |
| Token超限 | 截断历史 | AI只能理解部分历史 |
| 功能禁用 | 无历史模式 | 完全禁用历史功能 |

## 监控和告警建议

建议监控以下指标：

1. **历史消息获取成功率**
   - 计算公式：成功次数 / 总请求次数
   - 告警阈值：< 95%

2. **平均历史消息数量**
   - 计算公式：历史消息总数 / 成功次数
   - 用于评估用户对话长度

3. **Token截断频率**
   - 计算公式：截断次数 / 总请求次数
   - 告警阈值：> 20%

4. **降级事件频率**
   - 计算公式：降级次数 / 总请求次数
   - 告警阈值：> 5%

## 日志记录

所有错误和降级事件都会记录到日志，格式如下：

```
[历史上下文] 会话xxx: 历史上下文功能已禁用
[历史上下文] 获取历史消息失败: error, 使用降级方案
[历史上下文] 会话xxx无历史消息
[历史截断] 达到token限制(4000), 保留10条消息
[历史上下文] 会话xxx: 包含10条历史消息, 构建后prompt长度500字符
```

## 后续优化建议

1. **添加Redis缓存**
   - 缓存历史消息，减少数据库查询
   - TTL设置为5分钟

2. **添加熔断机制**
   - 当数据库查询失败率超过阈值时，自动禁用历史功能一段时间

3. **添加限流机制**
   - 对历史消息查询添加限流，防止数据库过载

4. **添加精确token计算**
   - 使用tokenizer库精确计算token数，避免估算误差