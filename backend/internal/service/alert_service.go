package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"
	"people-counting/internal/domain/service"
)

// AlertServiceImpl implements service.AlertService
type AlertServiceImpl struct {
	alertRepository     repository.AlertRepository
	alertTypeRepository repository.AlertTypeRepository
	cameraRepository    repository.CameraRepository
}

// NewAlertService creates a new alert service
func NewAlertService(
	alertRepository repository.AlertRepository,
	alertTypeRepository repository.AlertTypeRepository,
	cameraRepository repository.CameraRepository,
) service.AlertService {
	return &AlertServiceImpl{
		alertRepository:     alertRepository,
		alertTypeRepository: alertTypeRepository,
		cameraRepository:    cameraRepository,
	}
}

// GetAllAlerts retrieves paginated alert records with filters
func (s *AlertServiceImpl) GetAllAlerts(ctx context.Context, page, limit int, isActive, alertTypeID, cameraID, from, to, search, severity, status string, includeRelations bool) ([]entity.Alert, int64, error) {
	// Use default pagination values if invalid
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 50
	}

	// Prepare filters
	filters := make(map[string]interface{})

	// Add is_active filter if provided
	switch isActive {
	case "true":
		filters["is_active"] = true
	case "false":
		filters["is_active"] = false
	}

	// Add status filter if provided (maps to is_active and resolved_at logic)
	switch status {
	case "active":
		filters["is_active"] = true
		filters["resolved_at_null"] = true
	case "acknowledged":
		filters["is_active"] = false
		filters["resolved_at_null"] = true
	case "resolved":
		filters["resolved_at_not_null"] = true
	}

	// Add severity filter if provided
	if severity != "" {
		filters["severity"] = severity
	}

	// Add search filter if provided
	if search != "" {
		filters["search"] = search
	}

	// Add alert_type_id filter if provided
	if alertTypeID != "" {
		id, err := strconv.ParseUint(alertTypeID, 10, 64)
		if err != nil {
			return nil, 0, errors.New("invalid alert type ID")
		}

		// Check if alert type exists
		_, err = s.alertTypeRepository.FindByID(ctx, uint(id))
		if err != nil {
			return nil, 0, errors.New("alert type not found")
		}

		filters["alert_type_id"] = uint(id)
	}

	// Add camera_id filter if provided
	if cameraID != "" {
		id, err := strconv.ParseUint(cameraID, 10, 64)
		if err != nil {
			return nil, 0, errors.New("invalid camera ID")
		}

		// Check if camera exists
		_, err = s.cameraRepository.FindByID(ctx, uint(id))
		if err != nil {
			return nil, 0, errors.New("camera not found")
		}

		filters["camera_id"] = uint(id)
	}

	// Add time range filters if provided
	if from != "" {
		fromTime, err := parseFlexibleDate(from)
		if err != nil {
			return nil, 0, errors.New("invalid 'from' date format. " + err.Error())
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := parseFlexibleDate(to)
		if err != nil {
			return nil, 0, errors.New("invalid 'to' date format. " + err.Error())
		}
		filters["to"] = toTime
	}

	// Add include_relations flag if requested
	if includeRelations {
		filters["include_relations"] = true
	}

	return s.alertRepository.FindAll(ctx, page, limit, filters)
}

// GetActiveAlerts retrieves currently active alerts with pagination
func (s *AlertServiceImpl) GetActiveAlerts(ctx context.Context, page, limit int, alertTypeID, cameraID, from, to string, includeRelations bool) ([]entity.Alert, int64, int64, error) {
	// Use default pagination values if invalid
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 50
	}

	// Prepare filters
	filters := make(map[string]interface{})

	// Add alert_type_id filter if provided
	if alertTypeID != "" {
		id, err := strconv.ParseUint(alertTypeID, 10, 64)
		if err != nil {
			return nil, 0, 0, errors.New("invalid alert type ID")
		}

		// Check if alert type exists
		_, err = s.alertTypeRepository.FindByID(ctx, uint(id))
		if err != nil {
			return nil, 0, 0, errors.New("alert type not found")
		}

		filters["alert_type_id"] = uint(id)
	}

	// Add camera_id filter if provided
	if cameraID != "" {
		id, err := strconv.ParseUint(cameraID, 10, 64)
		if err != nil {
			return nil, 0, 0, errors.New("invalid camera ID")
		}

		// Check if camera exists
		_, err = s.cameraRepository.FindByID(ctx, uint(id))
		if err != nil {
			return nil, 0, 0, errors.New("camera not found")
		}

		filters["camera_id"] = uint(id)
	}

	// Add time range filters if provided
	if from != "" {
		fromTime, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return nil, 0, 0, errors.New("invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)")
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return nil, 0, 0, errors.New("invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)")
		}
		filters["to"] = toTime
	}

	// Add include_relations flag if requested
	if includeRelations {
		filters["include_relations"] = true
	}

	return s.alertRepository.FindActive(ctx, page, limit, filters)
}

// CreateAlert creates a new alert
func (s *AlertServiceImpl) CreateAlert(ctx context.Context, alert *entity.Alert) error {
	// Validate alert
	if alert.AlertTypeID == 0 {
		return errors.New("alert type ID is required")
	}

	if alert.Message == "" {
		return errors.New("message is required")
	}

	// Verify alert type exists
	_, err := s.alertTypeRepository.FindByID(ctx, alert.AlertTypeID)
	if err != nil {
		return errors.New("alert type not found")
	}

	// Verify camera exists if provided
	if alert.CameraID != 0 {
		_, err := s.cameraRepository.FindByID(ctx, alert.CameraID)
		if err != nil {
			return errors.New("camera not found")
		}
	}

	// Create alert
	return s.alertRepository.Create(ctx, alert)
}

// UpdateAlert updates an existing alert
func (s *AlertServiceImpl) UpdateAlert(ctx context.Context, alert *entity.Alert) error {
	// Validate alert
	if alert.ID == "" {
		return errors.New("alert ID is required")
	}

	if alert.AlertTypeID == 0 {
		return errors.New("alert type ID is required")
	}

	if alert.Message == "" {
		return errors.New("message is required")
	}

	// Verify alert type exists
	_, err := s.alertTypeRepository.FindByID(ctx, alert.AlertTypeID)
	if err != nil {
		return errors.New("alert type not found")
	}

	// Verify camera exists if camera_id is provided
	if alert.CameraID != 0 {
		_, err := s.cameraRepository.FindByID(ctx, alert.CameraID)
		if err != nil {
			return errors.New("camera not found")
		}
	}

	return s.alertRepository.Update(ctx, alert)
}

// ResolveAlert resolves (deactivates) an alert
func (s *AlertServiceImpl) ResolveAlert(ctx context.Context, id, resolvedBy, resolutionNote string) error {
	if id == "" {
		return errors.New("alert ID is required")
	}

	if resolvedBy == "" {
		return errors.New("resolved by is required")
	}

	return s.alertRepository.Resolve(ctx, id, resolvedBy, resolutionNote)
}

func (s *AlertServiceImpl) GetAlertByID(ctx context.Context, id string) (*entity.Alert, error) {
	if id == "" {
		return nil, errors.New("alert ID is required")
	}

	return s.alertRepository.FindByID(ctx, id)
}
