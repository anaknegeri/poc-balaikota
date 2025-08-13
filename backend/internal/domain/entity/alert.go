package entity

import (
	"time"
)

// Alert model represents alerts triggered in the system
type Alert struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AlertTypeID    uint       `gorm:"column:alert_type_id" json:"alert_type_id"`
	CameraID       uint       `gorm:"column:camera_id" json:"camera_id"`
	Message        string     `gorm:"type:text;not null;column:message" json:"message"`
	Severity       string     `gorm:"size:20;not null;column:severity" json:"severity"`
	ImageURL       string     `gorm:"size:255;column:image_url" json:"image_url"`
	IsActive       bool       `gorm:"default:true;column:is_active" json:"is_active"`
	DetectedAt     time.Time  `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP;column:detected_at" json:"detected_at"`
	ResolvedAt     *time.Time `gorm:"type:timestamp with time zone;column:resolved_at" json:"resolved_at"`
	ResolvedBy     string     `gorm:"size:100;column:resolved_by" json:"resolved_by"`
	ResolutionNote string     `gorm:"type:text;column:resolution_note" json:"resolution_note"`
	CreatedAt      time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`

	// Relationships
	AlertType AlertType `gorm:"foreignKey:AlertTypeID" json:"alert_type,omitempty"`
	Camera    *Camera   `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
}

// TableName returns the table name for the Alert model
func (Alert) TableName() string {
	return "alerts"
}
