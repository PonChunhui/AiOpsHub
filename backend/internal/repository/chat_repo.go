package repository

import (
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
)

// ChatRepository 对话历史仓库
// 提供对话会话和消息的数据库操作
type ChatRepository struct{}

// NewChatRepository 创建对话仓库实例
func NewChatRepository() *ChatRepository {
	return &ChatRepository{}
}

// CreateSession 创建新的对话会话
// 参数：session - 会话信息
// 返回：创建的会话对象和错误信息
func (r *ChatRepository) CreateSession(session *model.ChatSession) error {
	if session.ID == "" {
		session.ID = model.GenerateChatID()
	}
	return database.DB.Create(session).Error
}

// GetSessionByID 根据ID获取对话会话
// 参数：sessionID - 会话ID
// 返回：会话对象和错误信息
func (r *ChatRepository) GetSessionByID(sessionID string) (*model.ChatSession, error) {
	var session model.ChatSession
	err := database.DB.Where("id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionsByUserID 获取用户的所有对话会话
// 参数：userID - 用户ID，limit - 返回数量限制
// 返回：会话列表和错误信息
func (r *ChatRepository) GetSessionsByUserID(userID string, limit int) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	query := database.DB.Where("user_id = ?", userID).Order("updated_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&sessions).Error
	return sessions, err
}

// UpdateSession 更新对话会话信息
// 参数：session - 会话信息
// 返回：错误信息
func (r *ChatRepository) UpdateSession(session *model.ChatSession) error {
	return database.DB.Save(session).Error
}

// DeleteSession 删除对话会话及其所有消息
// 参数：sessionID - 会话ID
// 返回：错误信息
func (r *ChatRepository) DeleteSession(sessionID string) error {
	// 先删除会话的所有消息
	if err := database.DB.Where("session_id = ?", sessionID).Delete(&model.ChatMessage{}).Error; err != nil {
		return fmt.Errorf("删除会话消息失败: %w", err)
	}
	// 再删除会话
	if err := database.DB.Where("id = ?", sessionID).Delete(&model.ChatSession{}).Error; err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}
	return nil
}

// CreateMessage 创建新的对话消息
// 参数：message - 消息信息
// 返回：错误信息
func (r *ChatRepository) CreateMessage(message *model.ChatMessage) error {
	if message.ID == "" {
		message.ID = model.GenerateChatID()
	}
	return database.DB.Create(message).Error
}

// GetMessagesBySessionID 获取会话的所有消息
// 参数：sessionID - 会话ID
// 返回：消息列表和错误信息
func (r *ChatRepository) GetMessagesBySessionID(sessionID string) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := database.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// GetRecentMessages 获取会话最近的N条消息（用于构建上下文）
// 参数：sessionID - 会话ID，limit - 消息数量限制
// 返回：消息列表和错误信息
func (r *ChatRepository) GetRecentMessages(sessionID string, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := database.DB.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	// 需要反转顺序，使消息按时间正序排列
	if err == nil && len(messages) > 0 {
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}
	}
	return messages, err
}

// DeleteMessagesBySessionID 删除会话的所有消息
// 参数：sessionID - 会话ID
// 返回：错误信息
func (r *ChatRepository) DeleteMessagesBySessionID(sessionID string) error {
	return database.DB.Where("session_id = ?", sessionID).Delete(&model.ChatMessage{}).Error
}
