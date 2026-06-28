package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

// AgentRouter Agent 智能路由服务
// 根据用户问题智能选择合适的 Agent
type AgentRouter struct {
	agentSvc *AgentService
	llm      interface{} // LLM 接口用于智能路由
}

// NewAgentRouter 创建 Agent 路由器
func NewAgentRouter(agentSvc *AgentService) *AgentRouter {
	return &AgentRouter{
		agentSvc: agentSvc,
	}
}

// RouteAgent 智能路由选择 Agent
// 返回最适合处理用户问题的 Agent
func (r *AgentRouter) RouteAgent(ctx context.Context, userMessage string) (*model.Agent, error) {
	logger.Info(fmt.Sprintf("=== Agent Router: 分析用户问题 ==="))
	logger.Info(fmt.Sprintf("用户消息: %s", truncateContent(userMessage, 100)))

	// 第一步：规则路由（快速匹配）
	agent := r.routeByRules(userMessage)
	if agent != nil {
		logger.Info(fmt.Sprintf("✅ 规则路由匹配成功: %s (%s)", agent.Name, agent.ID))
		return agent, nil
	}

	// 第二步：LLM 智能路由（精确匹配）
	agent = r.routeByLLM(ctx, userMessage)
	if agent != nil {
		logger.Info(fmt.Sprintf("✅ LLM 路由匹配成功: %s (%s)", agent.Name, agent.ID))
		return agent, nil
	}

	// 如果没有匹配的 Agent，返回默认 Agent（或 nil 表示不使用 Agent）
	logger.Info("⚠️ 未找到匹配的 Agent，使用默认模式")
	return nil, nil
}

// routeByRules 规则路由（基于关键词匹配）
func (r *AgentRouter) routeByRules(userMessage string) *model.Agent {
	messageLower := strings.ToLower(userMessage)

	// 获取所有启用的 Agent
	agents, err := r.agentSvc.ListEnabled()
	if err != nil {
		logger.Error(fmt.Sprintf("获取 Agent 列表失败: %v", err))
		return nil
	}

	// 定义关键词映射规则（按优先级排序）
	// 格式：agentID -> {关键词列表，权重}
	type RuleConfig struct {
		keywords []string
		weight   int // 权重越高，优先级越高
	}

	rules := map[string]RuleConfig{
		"preset-alert-handler": {
			keywords: []string{
				"告警", "报警", "alert", "警报", "预警", "监控告警",
				"严重程度", "告警处理", "告警分析", "告警级别",
			},
			weight: 10, // 告警处理权重最高
		},
		"preset-fault-diagnosis": {
			keywords: []string{
				"故障", "诊断", "fault", "故障排查", "故障定位",
				"根本原因", "故障分析", "问题定位", "排错",
				"服务异常", "系统故障", "应用故障",
			},
			weight: 9,
		},
		"preset-log-analyzer": {
			keywords: []string{
				"日志", "log", "日志分析", "错误日志", "异常日志",
				"日志查询", "日志排查", "日志检查", "error log",
				"日志模式", "日志错误",
			},
			weight: 8,
		},
		"preset-change-executor": {
			keywords: []string{
				"变更", "change", "发布", "部署", "升级",
				"回滚", "变更执行", "配置变更", "版本发布",
				"变更方案", "变更计划",
			},
			weight: 7,
		},
		"preset-system-inspection": {
			keywords: []string{
				"巡检", "检查", "inspection", "健康检查", "系统巡检",
				"资源使用", "系统资源",
				"健康状态", "系统状态", "性能检查",
			},
			weight: 6,
		},
		"preset-doc-generator": {
			keywords: []string{
				"文档", "报告", "document", "report", "生成文档",
				"总结", "执行摘要", "操作文档", "运维文档",
				"生成报告", "结果报告",
			},
			weight: 5,
		},
		"preset-compliance-checker": {
			keywords: []string{
				"合规", "compliance", "配置合规",
				"安全检查", "规范检查", "合规性", "合规检查",
				"最佳实践",
			},
			weight: 4,
		},
		"preset-server-command": {
			keywords: []string{
				"命令", "执行", "command", "SSH", "服务器命令",
				"服务器操作", "shell", "bash",
				"服务器管理", "远程命令",
			},
			weight: 3,
		},
		"preset-auto-inspection": {
			keywords: []string{
				"批量巡检", "自动巡检", "多服务器", "批量检查",
				"自动检查", "巡检报告", "综合巡检", "批量诊断",
			},
			weight: 2,
		},
	}

	// 特殊关键词（单独处理，避免冲突）
	specialKeywords := map[string]string{
		"CPU": "preset-system-inspection",
		"内存":  "preset-system-inspection",
		"磁盘":  "preset-system-inspection",
	}

	// 匹配关键词并计算得分
	type Match struct {
		agent *model.Agent
		score int
		count int
	}

	matches := []Match{}

	for _, agent := range agents {
		if ruleConfig, ok := rules[agent.ID]; ok {
			matchCount := 0
			for _, keyword := range ruleConfig.keywords {
				if strings.Contains(messageLower, strings.ToLower(keyword)) {
					matchCount++
				}
			}

			if matchCount > 0 {
				// 计算得分：匹配数量 * 权重
				score := matchCount * ruleConfig.weight
				matches = append(matches, Match{
					agent: &agent,
					score: score,
					count: matchCount,
				})
				logger.Info(fmt.Sprintf("  Agent %s 匹配 %d 个关键词，得分: %d", agent.Name, matchCount, score))
			}
		}
	}

	// 检查特殊关键词（只有在没有其他匹配时才使用）
	if len(matches) == 0 {
		for keyword, agentID := range specialKeywords {
			if strings.Contains(userMessage, keyword) {
				for _, agent := range agents {
					if agent.ID == agentID {
						logger.Info(fmt.Sprintf("  特殊关键词 '%s' 匹配 Agent: %s", keyword, agent.Name))
						return &agent
					}
				}
			}
		}
	}

	// 选择得分最高的 Agent
	if len(matches) > 0 {
		bestMatch := matches[0]
		for _, match := range matches {
			if match.score > bestMatch.score {
				bestMatch = match
			}
		}
		logger.Info(fmt.Sprintf("  最佳匹配: %s (得分: %d, 匹配关键词: %d)", bestMatch.agent.Name, bestMatch.score, bestMatch.count))
		return bestMatch.agent
	}

	return nil
}

// routeByLLM LLM 智能路由（基于意图理解）
// 让 LLM 分析用户问题并选择最合适的 Agent
func (r *AgentRouter) routeByLLM(ctx context.Context, userMessage string) *model.Agent {
	// 获取所有启用的 Agent
	agents, err := r.agentSvc.ListEnabled()
	if err != nil {
		logger.Error(fmt.Sprintf("获取 Agent 列表失败: %v", err))
		return nil
	}

	if len(agents) == 0 {
		return nil
	}

	// 构建 Agent 选择 Prompt
	var agentListStr strings.Builder
	agentListStr.WriteString("可选的 Agent:\n")
	for i, agent := range agents {
		agentListStr.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", i+1, agent.Name, agent.ID, agent.Description))
	}

	selectionPrompt := fmt.Sprintf(`
你是一个智能路由助手，需要根据用户问题选择最合适的 Agent 来处理。

%s

用户问题: %s

请分析用户问题的意图和需求，选择最合适的 Agent。
只需返回 Agent 的 ID（例如：preset-alert-handler），不要返回其他内容。

如果问题不适合任何 Agent，返回 "none"。

Agent ID:
`, agentListStr.String(), userMessage)

	// TODO: 这里需要调用 LLM 进行智能选择
	// 目前先用简单的关键词匹配作为演示
	// 实际使用时需要调用 ChatService 的 LLM

	logger.Info("LLM 路由提示: " + truncateContent(selectionPrompt, 200))

	// 暂时返回 nil，等后续集成 LLM
	return nil
}

// GetAgentSystemPrompt 获取 Agent 的 SystemPrompt
func (r *AgentRouter) GetAgentSystemPrompt(agentID string) string {
	agent, err := r.agentSvc.GetByID(agentID)
	if err != nil {
		logger.Error(fmt.Sprintf("获取 Agent 失败: %v", err))
		return ""
	}

	return agent.SystemPrompt
}
