package postgres

import (
	"context"
	"errors"
	"strings"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"

	"gorm.io/gorm"
)

// AlertTypeRepositoryImpl implements repository.AlertTypeRepository
type AlertTypeRepositoryImpl struct {
	db *gorm.DB
}

// NewAlertTypeRepository creates a new alert type repository
func NewAlertTypeRepository(db *gorm.DB) repository.AlertTypeRepository {
	return &AlertTypeRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves all alert types
func (r *AlertTypeRepositoryImpl) FindAll(ctx context.Context) ([]entity.AlertType, error) {
	var alertTypes []entity.AlertType

	result := r.db.WithContext(ctx).Order("id ASC").Find(&alertTypes)
	if result.Error != nil {
		return nil, result.Error
	}

	return alertTypes, nil
}

// FindByID finds an alert type by its ID
func (r *AlertTypeRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.AlertType, error) {
	var alertType entity.AlertType

	result := r.db.WithContext(ctx).First(&alertType, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert type not found")
		}
		return nil, result.Error
	}

	return &alertType, nil
}

// Create adds a new alert type to the database
func (r *AlertTypeRepositoryImpl) Create(ctx context.Context, alertType *entity.AlertType) (*entity.AlertType, error) {
	if err := r.db.WithContext(ctx).Create(alertType).Error; err != nil {
		return nil, err
	}
	return alertType, nil
}

func (r *AlertTypeRepositoryImpl) FindByName(ctx context.Context, typeName string) (*entity.AlertType, error) {
	var alertType *entity.AlertType

	result := r.db.WithContext(ctx).
		Where("LOWER(name) = LOWER(?)", strings.ToLower(typeName)).
		First(&alertType)

	if result.Error != nil {
		return nil, result.Error
	}

	return alertType, nil
}
