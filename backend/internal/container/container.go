package container

import (
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/internal/service"
)

type Container struct {
	Repositories *RepositoryContainer
	Services     *ServiceContainer
}

type RepositoryContainer struct {
	Agent *repository.AgentRepository
	Tool  *repository.ToolRepository
	Alert *repository.AlertRepository
	User  *repository.UserRepository
	Chat  *repository.ChatRepository
	MCP   *repository.MCPRepository
	RAG   *repository.RAGRepository
	Token *repository.TokenRepository
}

type ServiceContainer struct {
	Agent       *service.AgentService
	Tool        *service.ToolService
	Alert       *service.AlertService
	User        *service.UserService
	Chat        *service.ChatService
	MCP         *service.MCPService
	RAG         *service.RAGService
	Token       *service.TokenService
	AgentRouter *service.AgentRouter
}

func NewContainer() *Container {
	repos := initRepositories()
	services := initServices(repos)

	return &Container{
		Repositories: repos,
		Services:     services,
	}
}

func initRepositories() *RepositoryContainer {
	return &RepositoryContainer{
		Agent: repository.NewAgentRepository(),
		Tool:  repository.NewToolRepository(),
		Alert: repository.NewAlertRepository(),
		User:  repository.NewUserRepository(),
		Chat:  repository.NewChatRepository(),
		MCP:   repository.NewMCPRepository(),
		RAG:   repository.NewRAGRepository(),
		Token: repository.NewTokenRepository(nil),
	}
}

func initServices(repos *RepositoryContainer) *ServiceContainer {
	agentSvc := service.NewAgentService()
	toolSvc := service.NewToolService(repos.Tool)
	alertSvc := service.NewAlertService()
	userSvc := service.NewUserService()
	mcpSvc := service.NewMCPService()
	ragSvc := service.NewRAGService("aiops_knowledge")
	tokenSvc := service.NewTokenServiceWithRepo(repos.Token)

	agentRouter := service.NewAgentRouter(agentSvc)

	return &ServiceContainer{
		Agent:       agentSvc,
		Tool:        toolSvc,
		Alert:       alertSvc,
		User:        userSvc,
		Chat:        nil,
		MCP:         mcpSvc,
		RAG:         ragSvc,
		Token:       tokenSvc,
		AgentRouter: agentRouter,
	}
}
