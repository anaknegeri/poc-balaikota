package postgres

import (
	"context"
	"errors"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"

	"gorm.io/gorm"
)

// CameraRepositoryImpl implements repository.CameraRepository
type CameraRepositoryImpl struct {
	db *gorm.DB
}

// NewCameraRepository creates a new camera repository
func NewCameraRepository(db *gorm.DB) repository.CameraRepository {
	return &CameraRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves cameras with optional filters
func (r *CameraRepositoryImpl) FindAll(ctx context.Context, filters map[string]interface{}) ([]entity.Camera, error) {
	var cameras []entity.Camera

	query := r.db.WithContext(ctx).Order("id ASC")

	// Apply filters if provided
	if filters != nil {
		if status, ok := filters["status"].(string); ok && status != "" {
			query = query.Where("status = ?", status)
		}
	}

	result := query.Find(&cameras)
	if result.Error != nil {
		return nil, result.Error
	}

	return cameras, nil
}

// FindByID finds a camera by its ID
func (r *CameraRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.Camera, error) {
	var camera entity.Camera

	result := r.db.WithContext(ctx).First(&camera, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("camera not found")
		}
		return nil, result.Error
	}

	return &camera, nil
}

// FindByArea finds cameras by area ID
func (r *CameraRepositoryImpl) FindByArea(ctx context.Context, areaID uint) ([]entity.Camera, error) {
	var cameras []entity.Camera

	result := r.db.WithContext(ctx).Where("area_id = ?", areaID).Find(&cameras)
	if result.Error != nil {
		return nil, result.Error
	}

	return cameras, nil
}

// Create adds a new camera to the database
func (r *CameraRepositoryImpl) Create(ctx context.Context, camera *entity.Camera) error {
	result := r.db.WithContext(ctx).Create(camera)
	return result.Error
}

// Update modifies an existing camera in the database
func (r *CameraRepositoryImpl) Update(ctx context.Context, camera *entity.Camera) error {
	// Check if camera exists
	var count int64
	r.db.WithContext(ctx).Model(&entity.Camera{}).Where("id = ?", camera.ID).Count(&count)
	if count == 0 {
		return errors.New("camera not found")
	}

	result := r.db.WithContext(ctx).Save(camera)
	return result.Error
}

// UpdateStatus updates just the status of a camera
func (r *CameraRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	// Check if camera exists
	var count int64
	r.db.WithContext(ctx).Model(&entity.Camera{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return errors.New("camera not found")
	}

	// Update status and last online time
	result := r.db.WithContext(ctx).Model(&entity.Camera{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"last_online": time.Now(),
			"updated_at":  time.Now(),
		})

	return result.Error
}

// Delete removes a camera from the database
func (r *CameraRepositoryImpl) Delete(ctx context.Context, id uint) error {
	// Check if camera exists
	var count int64
	r.db.WithContext(ctx).Model(&entity.Camera{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return errors.New("camera not found")
	}

	result := r.db.WithContext(ctx).Delete(&entity.Camera{}, id)
	return result.Error
}
