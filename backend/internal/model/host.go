package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HostGroup struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid"`
	Name        string         `json:"name" gorm:"not null;type:varchar(100)"`
	ParentID    *string        `json:"parent_id,omitempty" gorm:"type:uuid;index"`
	Description string         `json:"description" gorm:"type:text"`
	Level       int            `json:"level" gorm:"default:0"`
	Path        string         `json:"path" gorm:"type:text"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100)"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"index"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Children []HostGroup `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Hosts    []Host      `json:"hosts,omitempty" gorm:"foreignKey:GroupID"`
}

type Host struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid"`
	GroupID     string         `json:"group_id" gorm:"type:uuid;index"`
	Name        string         `json:"name" gorm:"not null;type:varchar(100)"`
	HostType    string         `json:"host_type" gorm:"type:varchar(20);default:'linux'"`
	IP          string         `json:"ip" gorm:"not null;type:varchar(50);index"`
	Port        int            `json:"port" gorm:"default:22"`
	Username    string         `json:"username" gorm:"not null;type:varchar(50)"`
	AuthType    string         `json:"auth_type" gorm:"type:varchar(20);default:'password'"`
	Password    string         `json:"-" gorm:"type:text"`
	PrivateKey  string         `json:"-" gorm:"type:text"`
	PublicKey   string         `json:"public_key" gorm:"type:text"`
	Remark      string         `json:"remark" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:varchar(20);default:'active'"`
	LastCheckAt time.Time      `json:"last_check_at"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100)"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"index"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Group HostGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
}

type SSHSessionLog struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	HostID    string    `json:"host_id" gorm:"type:uuid;index"`
	UserID    string    `json:"user_id" gorm:"type:uuid;index"`
	Action    string    `json:"action" gorm:"type:varchar(50)"`
	Command   string    `json:"command" gorm:"type:text"`
	SessionID string    `json:"session_id" gorm:"type:uuid;index"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int       `json:"duration"`
	Result    string    `json:"result" gorm:"type:text"`
	IPAddress string    `json:"ip_address" gorm:"type:varchar(50)"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	Host Host `json:"host,omitempty" gorm:"foreignKey:HostID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func NewHostGroup() *HostGroup {
	return &HostGroup{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewHost() *Host {
	return &Host{
		ID:        uuid.New().String(),
		Port:      22,
		HostType:  "linux",
		AuthType:  "password",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewSSHSessionLog() *SSHSessionLog {
	return &SSHSessionLog{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
	}
}
