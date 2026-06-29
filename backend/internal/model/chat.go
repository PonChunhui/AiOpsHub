package model

import (
	"time"

	"github.com/google/uuid"
)

// ChatSession 对话会话模型
// 用于记录用户的对话会话信息
type ChatSession struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`           // 会话ID
	UserID    string    `json:"user_id" gorm:"not null;type:varchar(36);index"`  // 用户ID
	Title     string    `json:"title" gorm:"type:varchar(255)"`                  // 会话标题
	Model     string    `json:"model" gorm:"type:varchar(50)"`                   // 使用的模型名称
	Status    string    `json:"status" gorm:"type:varchar(20);default:'active'"` // 会话状态：active, archived
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`                // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`                // 更新时间
}

// ChatMessage 对话消息模型
// 用于记录对话中的每一条消息
type ChatMessage struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(36)"`             // 消息ID
	SessionID     string    `json:"session_id" gorm:"not null;type:varchar(36);index"` // 会话ID
	Role          string    `json:"role" gorm:"not null;type:varchar(20)"`             // 角色：user, assistant, system
	Content       string    `json:"content" gorm:"type:text"`                          // 消息内容（支持markdown）
	Tokens        int       `json:"tokens" gorm:"default:0"`                           // Token数量
	RAGReferences string    `json:"rag_references" gorm:"type:text"`                   // RAG引用的JSON数据
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime;index"`            // 创建时间
}

// ChatMessageResponse 用于API响应的消息结构
// RAGReferences字段返回数组而不是JSON字符串
type ChatMessageResponse struct {
	ID            string                   `json:"id"`
	SessionID     string                   `json:"session_id"`
	Role          string                   `json:"role"`
	Content       string                   `json:"content"`
	Tokens        int                      `json:"tokens"`
	RAGReferences []map[string]interface{} `json:"rag_references"`
	CreatedAt     time.Time                `json:"created_at"`
}

// ChatSessionWithMessages 对话会话包含消息的结构
// 用于返回包含消息列表的会话信息
type ChatSessionWithMessages struct {
	ChatSession
	Messages []ChatMessage `json:"messages"` // 会话中的消息列表
}

// ChatSessionWithMessagesResponse 用于API响应的会话结构
// Messages字段使用ChatMessageResponse类型
type ChatSessionWithMessagesResponse struct {
	ChatSession
	Messages []ChatMessageResponse `json:"messages"` // 会话中的消息列表
}

// GenerateChatID 生成对话ID
// 返回一个唯一的UUID字符串
func GenerateChatID() string {
	return uuid.New().String()
}
