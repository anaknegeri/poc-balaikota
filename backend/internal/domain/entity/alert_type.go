package entity

import (
	"time"
)

// AlertType model represents types of alerts that can be triggered
type AlertType struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"size:50;not null;column:name" json:"name"`
	DisplayName string    `gorm:"size:50;not null;column:display_name" json:"display_name"`
	Icon        string    `gorm:"size:50;column:icon" json:"icon"`
	Color       string    `gorm:"size:20;column:color" json:"color"`
	Description string    `gorm:"type:text;column:description" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`

	// Relationships
	Alerts []Alert `gorm:"foreignKey:AlertTypeID" json:"alerts,omitempty"`
}

// TableName returns the table name for the AlertType model
func (AlertType) TableName() string {
	return "alert_types"
}
