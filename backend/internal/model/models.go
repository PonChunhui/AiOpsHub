package model

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null;type:varchar(100)"`
	Type         string    `json:"type" gorm:"type:varchar(50);default:'preset'"` // Legacy field, will be deprecated
	Avatar       string    `json:"avatar" gorm:"type:varchar(10)"`                // Emoji 头像
	Role         string    `json:"role" gorm:"type:varchar(200)"`                 // 角色描述
	Category     string    `json:"category" gorm:"type:varchar(100)"`             // 分类
	Description  string    `json:"description" gorm:"type:text"`                  // 功能描述
	SystemPrompt string    `json:"system_prompt" gorm:"type:text"`                // 系统提示词
	Model        string    `json:"model" gorm:"type:varchar(50)"`                 // 绑定的 LLM 模型
	Temperature  float64   `json:"temperature" gorm:"type:decimal(3,2)"`          // 温度参数
	IsPreset     bool      `json:"is_preset" gorm:"default:false"`                // 是否预设
	Enabled      bool      `json:"enabled" gorm:"default:true"`                   // 是否启用
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"index"`
}

type Alert struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Source      string    `json:"source" gorm:"not null"`
	Severity    string    `json:"severity" gorm:"not null"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"default:'open'"`
	RawData     string    `json:"raw_data" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RAGDocument struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text"`
	DocType   string    `json:"doc_type" gorm:"type:varchar(50)"`  // 文档类型：sop / faq / alert
	Component string    `json:"component" gorm:"type:varchar(50)"` // 组件名：mysql / k8s / redis
	Tags      string    `json:"tags" gorm:"type:varchar(500)"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy string    `json:"updated_by" gorm:"type:varchar(100)"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index"`
}

type Datasource struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"`
	Config    string    `json:"config" gorm:"type:text"`
	Status    string    `json:"status" gorm:"default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Role      string    `json:"role" gorm:"default:'user'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tool struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"`
	Description string    `json:"description"`
	Config      string    `json:"config" gorm:"type:text"`
	Status      string    `json:"status" gorm:"default:'active'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func GenerateID() string {
	return uuid.New().String()
}

type MCPServer struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	URL         string    `json:"url" gorm:"not null"`
	AuthType    string    `json:"auth_type" gorm:"type:varchar(20)"` // api_key, bearer, none
	AuthToken   string    `json:"-" gorm:"type:text"`                // 加密存储，不返回给前端
	Status      string    `json:"status" gorm:"default:'active'"`    // active, inactive
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:varchar(100)"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"index"`
}

type TokenUsageRecord struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	SessionID    string    `json:"session_id" gorm:"type:varchar(36);index"`
	AgentID      string    `json:"agent_id" gorm:"type:varchar(36);index"`
	Model        string    `json:"model" gorm:"type:varchar(50)"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	Cost         float64   `json:"cost"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}
