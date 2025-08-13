package entity

import "time"

type FaceRecognition struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CameraID   uint      `gorm:"not null;column:camera_id" json:"camera_id"`
	ImagePath  string    `gorm:"size:255;column:image_path" json:"image_path"`
	ObjectName string    `gorm:"size:255;column:object_name" json:"object_name"`
	DetectedAt time.Time `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP;column:detected_at" json:"detected_at"`
	CreatedAt  time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`

	// Relationships
	Camera *Camera `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
}

// TableName returns the table name for the Alert model
func (FaceRecognition) TableName() string {
	return "face_recognitions"
}
