package service

import (
	"context"
	"errors"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"
	"people-counting/internal/domain/service"
)

// AlertTypeServiceImpl implements service.AlertTypeService
type AlertTypeServiceImpl struct {
	alertTypeRepository repository.AlertTypeRepository
}

// NewAlertTypeService creates a new alert type service
func NewAlertTypeService(
	alertTypeRepository repository.AlertTypeRepository,
) service.AlertTypeService {
	return &AlertTypeServiceImpl{
		alertTypeRepository: alertTypeRepository,
	}
}

// GetAllAlertTypes retrieves all alert types
func (s *AlertTypeServiceImpl) GetAllAlertTypes(ctx context.Context) ([]entity.AlertType, error) {
	return s.alertTypeRepository.FindAll(ctx)
}

// CreateAlertType creates a new alert type
func (s *AlertTypeServiceImpl) CreateAlertType(ctx context.Context, alertType *entity.AlertType) (*entity.AlertType, error) {
	// Validate alert type
	if alertType.Name == "" {
		return nil, errors.New("name is required")
	}

	if alertType.Icon == "" {
		return nil, errors.New("icon is required")
	}

	if alertType.Color == "" {
		return nil, errors.New("color is required")
	}

	// Create alert type
	return s.alertTypeRepository.Create(ctx, alertType)
}

func (s *AlertTypeServiceImpl) GetAlertTypeByName(ctx context.Context, typeName string) (*entity.AlertType, error) {
	camera, err := s.alertTypeRepository.FindByName(ctx, typeName)
	if err != nil {
		return nil, err
	}

	return camera, nil
}
