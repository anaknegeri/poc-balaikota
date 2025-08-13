package entity

import (
	"time"
)

// Camera model represents a camera in the system
type Camera struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"size:100;not null;column:name" json:"name"`
	IPAddress string    `gorm:"size:15;column:ip_address" json:"ip_address"`
	Location  string    `gorm:"size:100;column:location" json:"location"`
	Status    string    `gorm:"size:20;default:active;column:status" json:"status"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`

	WsURL     string `gorm:"size:100;column:ws_url" json:"ws_url"`
	StreamURL string `gorm:"-" json:"stream_url,omitempty"`
	ImageURL  string `gorm:"-" json:"image_url,omitempty"`

	// Relationships
	PeopleCounts []PeopleCount `gorm:"foreignKey:CameraID" json:"people_counts,omitempty"`
	Alerts       []Alert       `gorm:"foreignKey:CameraID" json:"alerts,omitempty"`
}

// TableName returns the table name for the Camera model
func (Camera) TableName() string {
	return "cameras"
}
