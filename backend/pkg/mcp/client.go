package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type Client struct {
	URL        string
	AuthType   string
	AuthToken  string
	SessionID  string
	HTTPClient *http.Client
}

func NewClient(url, authType, authToken string) *Client {
	return &Client{
		URL:       url,
		AuthType:  authType,
		AuthToken: authToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Initialize(ctx context.Context) (*InitializeResult, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "AiOpsHub",
				"version": "1.0.0",
			},
		},
	}

	resp, httpResp, err := c.sendRequestWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}

	// 从响应 header 获取 session ID
	if sessionID := httpResp.Header.Get("mcp-session-id"); sessionID != "" {
		c.SessionID = sessionID
		logger.Info(fmt.Sprintf("MCP Session ID: %s", sessionID))
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	var result InitializeResult
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("MCP Server initialized: %s v%s", result.ServerInfo.Name, result.ServerInfo.Version))
	return &result, nil
}

func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	var result ToolListResult
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("MCP Server has %d tools", len(result.Tools)))
	return result.Tools, nil
}

func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolCallResult, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("call tool error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
	}

	var result ToolCallResult
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) sendRequest(ctx context.Context, req JSONRPCRequest) (*JSONRPCResponse, error) {
	resp, _, err := c.sendRequestWithResponse(ctx, req)
	return resp, err
}

func (c *Client) sendRequestWithResponse(ctx context.Context, req JSONRPCRequest) (*JSONRPCResponse, *http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if c.SessionID != "" {
		httpReq.Header.Set("mcp-session-id", c.SessionID)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, err
	}
	resp.Body.Close()

	// 尝试从 SSE 格式解析
	respStr := string(respBody)
	if strings.HasPrefix(respStr, "id:") || strings.HasPrefix(respStr, "event:") {
		// SSE 格式响应
		sessionID := ""
		dataLine := ""
		for _, line := range strings.Split(respStr, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "id:") {
				sessionID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			} else if strings.HasPrefix(line, "data:") {
				dataLine = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			}
		}

		if sessionID != "" && c.SessionID == "" {
			c.SessionID = sessionID
			logger.Info(fmt.Sprintf("MCP Session ID from SSE: %s", sessionID))
		}

		if dataLine != "" {
			var jsonResp JSONRPCResponse
			if err := json.Unmarshal([]byte(dataLine), &jsonResp); err != nil {
				return nil, nil, fmt.Errorf("JSON parse error from SSE: %w (data: %s)", err, dataLine)
			}
			return &jsonResp, resp, nil
		}

		return nil, nil, fmt.Errorf("SSE response missing data line")
	}

	// 普通 JSON 响应
	var jsonResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &jsonResp); err != nil {
		return nil, nil, fmt.Errorf("JSON parse error: %w (body: %s)", err, string(respBody))
	}

	return &jsonResp, resp, nil
}

func (c *Client) setAuthHeader(req *http.Request) {
	switch c.AuthType {
	case "api_key":
		req.Header.Set("X-API-Key", c.AuthToken)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	case "basic":
		// Basic Auth: username:password 或 username:api_token
		// 格式: Authorization: Basic base64(username:token)
		req.Header.Set("Authorization", "Basic "+c.AuthToken)
	}
}

func (c *Client) TestConnection(ctx context.Context) error {
	result, err := c.Initialize(ctx)
	if err != nil {
		return err
	}
	if result == nil || result.ServerInfo.Name == "" {
		return fmt.Errorf("invalid initialize response")
	}
	return nil
}

func (c *Client) GetToolSchema(ctx context.Context, toolName string) (map[string]interface{}, error) {
	tools, err := c.ListTools(ctx)
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if tool.Name == toolName {
			return tool.InputSchema, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", toolName)
}

func ExtractTextContent(result *ToolCallResult) string {
	var texts []string
	for _, block := range result.Content {
		if block.Type == "text" {
			texts = append(texts, block.Text)
		}
	}
	return strings.Join(texts, "\n")
}
