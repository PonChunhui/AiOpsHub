package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/mcp"
	"github.com/google/uuid"
)

type MCPService struct {
	repo    *repository.MCPRepository
	clients map[string]*mcp.Client
	mu      sync.RWMutex
}

func NewMCPService() *MCPService {
	return &MCPService{
		repo:    repository.NewMCPRepository(),
		clients: make(map[string]*mcp.Client),
	}
}

func (s *MCPService) Create(ctx context.Context, name, description, url, authType, authToken, createdBy string) (*model.MCPServer, error) {
	server := &model.MCPServer{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		URL:         url,
		AuthType:    authType,
		AuthToken:   authToken,
		Status:      "active",
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(server); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.clients[server.ID] = mcp.NewClient(url, authType, authToken)
	s.mu.Unlock()

	logger.Info(fmt.Sprintf("MCP Server created: %s (%s)", name, url))
	return server, nil
}

func (s *MCPService) GetByID(id string) (*model.MCPServer, error) {
	return s.repo.GetByID(id)
}

func (s *MCPService) List(page, pageSize int) ([]model.MCPServer, int64, error) {
	return s.repo.List(page, pageSize)
}

func (s *MCPService) Update(ctx context.Context, id string, name, description, url, authType, authToken, updatedBy string) (*model.MCPServer, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		server.Name = name
	}
	if description != "" {
		server.Description = description
	}
	if url != "" {
		server.URL = url
	}
	if authType != "" {
		server.AuthType = authType
	}
	if authToken != "" {
		server.AuthToken = authToken
	}
	server.UpdatedBy = updatedBy
	server.UpdatedAt = time.Now()

	if err := s.repo.Update(server); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.clients[id] = mcp.NewClient(server.URL, server.AuthType, server.AuthToken)
	s.mu.Unlock()

	logger.Info(fmt.Sprintf("MCP Server updated: %s", id))
	return server, nil
}

func (s *MCPService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()

	logger.Info(fmt.Sprintf("MCP Server deleted: %s", id))
	return nil
}

func (s *MCPService) TestConnection(ctx context.Context, id string) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	client := mcp.NewClient(server.URL, server.AuthType, server.AuthToken)
	return client.TestConnection(ctx)
}

func (s *MCPService) GetTools(ctx context.Context, id string) ([]mcp.Tool, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	client := s.getClient(server)

	// 确保先初始化获取 session ID
	if client.SessionID == "" {
		_, err := client.Initialize(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize MCP server: %w", err)
		}
	}

	return client.ListTools(ctx)
}

func (s *MCPService) getClient(server *model.MCPServer) *mcp.Client {
	s.mu.RLock()
	client, exists := s.clients[server.ID]
	s.mu.RUnlock()

	if exists {
		return client
	}

	client = mcp.NewClient(server.URL, server.AuthType, server.AuthToken)
	s.mu.Lock()
	s.clients[server.ID] = client
	s.mu.Unlock()

	return client
}

func (s *MCPService) GetAllActiveTools(ctx context.Context) (map[string][]mcp.Tool, error) {
	servers, err := s.repo.ListActive()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]mcp.Tool)
	for _, server := range servers {
		client := s.getClient(&server)

		// 确保先初始化获取 session ID
		if client.SessionID == "" {
			_, err := client.Initialize(ctx)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to initialize MCP server %s: %v", server.Name, err))
				continue
			}
		}

		tools, err := client.ListTools(ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to list tools from %s: %v", server.Name, err))
			continue
		}
		result[server.Name] = tools
	}

	return result, nil
}

func (s *MCPService) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
	server, err := s.repo.GetByID(serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	client := s.getClient(server)
	result, err := client.CallTool(ctx, toolName, arguments)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *MCPService) FindServerByToolName(ctx context.Context, toolName string) (string, error) {
	servers, err := s.repo.ListActive()
	if err != nil {
		return "", err
	}

	for _, server := range servers {
		client := s.getClient(&server)

		if client.SessionID == "" {
			_, err := client.Initialize(ctx)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to initialize MCP server %s: %v", server.Name, err))
				continue
			}
		}

		tools, err := client.ListTools(ctx)
		if err != nil {
			continue
		}
		for _, tool := range tools {
			if tool.Name == toolName {
				return server.ID, nil
			}
		}
	}

	return "", fmt.Errorf("tool %s not found in any active MCP server", toolName)
}
