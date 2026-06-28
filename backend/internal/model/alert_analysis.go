package model

import "time"

type AlertAnalysisResult struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	AlertID      string    `gorm:"type:varchar(255);index" json:"alert_id"`
	Status       string    `gorm:"type:varchar(50)" json:"status"`
	Result       string    `gorm:"type:text" json:"result"`
	AnalysisText string    `gorm:"type:text" json:"analysis_text"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AlertAnalysisResult) TableName() string {
	return "alert_analysis_results"
}
