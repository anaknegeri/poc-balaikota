package postgres

import (
	"context"
	"errors"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"

	"gorm.io/gorm"
)

// FaceRecognitionRepositoryImpl implements repository.FaceRecognitionRepository
type FaceRecognitionRepositoryImpl struct {
	db *gorm.DB
}

// NewFaceRecognitionRepository creates a new face recognition repository
func NewFaceRecognitionRepository(db *gorm.DB) repository.FaceRecognitionRepository {
	return &FaceRecognitionRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves paginated face recognition records with filters
func (r *FaceRecognitionRepositoryImpl) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.FaceRecognition, int64, error) {
	var faceRecognitions []entity.FaceRecognition
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Build query
	query := r.db.WithContext(ctx).Model(&entity.FaceRecognition{}).Order("detected_at DESC")

	// Apply filters if provided
	if filters != nil {
		if cameraID, ok := filters["camera_id"].(uint); ok && cameraID != 0 {
			query = query.Where("camera_id = ?", cameraID)
		}

		if from, ok := filters["from"].(time.Time); ok && !from.IsZero() {
			query = query.Where("detected_at >= ?", from)
		}

		if to, ok := filters["to"].(time.Time); ok && !to.IsZero() {
			query = query.Where("detected_at <= ?", to)
		}

		if includeRelations, ok := filters["include_relations"].(bool); ok && includeRelations {
			query = query.Preload("Camera")
		}
	}

	// Get total count for pagination
	countQuery := query
	countQuery.Count(&total)

	// Get paginated results
	result := query.Limit(limit).Offset(offset).Find(&faceRecognitions)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return faceRecognitions, total, nil
}

// FindByID finds a face recognition record by its ID
func (r *FaceRecognitionRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.FaceRecognition, error) {
	var faceRecognition entity.FaceRecognition

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&faceRecognition)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("face recognition not found")
		}
		return nil, result.Error
	}

	return &faceRecognition, nil
}

// Create adds a new face recognition record to the database
func (r *FaceRecognitionRepositoryImpl) Create(ctx context.Context, data *entity.FaceRecognition) error {
	// Set detected_at to current time if not provided
	if data.DetectedAt.IsZero() {
		data.DetectedAt = time.Now()
	}

	result := r.db.WithContext(ctx).Create(data)
	if result.Error != nil {
		return result.Error
	}

	// Load related entities if needed
	if data.ID != "" {
		r.db.WithContext(ctx).Preload("Camera").First(data, "id = ?", data.ID)
	}

	return nil
}

// Update updates an existing face recognition record in the database
func (r *FaceRecognitionRepositoryImpl) Update(ctx context.Context, data *entity.FaceRecognition) error {
	// Check if face recognition exists
	var existingData entity.FaceRecognition
	result := r.db.WithContext(ctx).Where("id = ?", data.ID).First(&existingData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("face recognition not found")
		}
		return result.Error
	}

	// Update the face recognition
	now := time.Now()
	data.UpdatedAt = now

	updateResult := r.db.WithContext(ctx).Model(&existingData).Updates(data)
	if updateResult.Error != nil {
		return updateResult.Error
	}

	// Load related entities if needed
	if data.ID != "" {
		r.db.WithContext(ctx).Preload("Camera").First(data, "id = ?", data.ID)
	}

	return nil
}
