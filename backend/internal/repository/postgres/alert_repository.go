package postgres

import (
	"context"
	"errors"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"

	"gorm.io/gorm"
)

// AlertRepositoryImpl implements repository.AlertRepository
type AlertRepositoryImpl struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *gorm.DB) repository.AlertRepository {
	return &AlertRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves paginated alert records with filters
func (r *AlertRepositoryImpl) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.Alert, int64, error) {
	var alerts []entity.Alert
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Build query
	query := r.db.WithContext(ctx).Model(&entity.Alert{}).Order("detected_at DESC")

	// Apply filters if provided
	if filters != nil {
		if isActive, ok := filters["is_active"].(bool); ok {
			query = query.Where("is_active = ?", isActive)
		}

		if alertTypeID, ok := filters["alert_type_id"].(uint); ok && alertTypeID != 0 {
			query = query.Where("alert_type_id = ?", alertTypeID)
		}

		if cameraID, ok := filters["camera_id"].(uint); ok && cameraID != 0 {
			query = query.Where("camera_id = ?", cameraID)
		}

		if severity, ok := filters["severity"].(string); ok && severity != "" {
			query = query.Where("severity = ?", severity)
		}

		if search, ok := filters["search"].(string); ok && search != "" {
			searchPattern := "%" + search + "%"
			query = query.Where("message ILIKE ? OR resolution_note ILIKE ?", searchPattern, searchPattern)
		}

		// Status-based filters
		if _, ok := filters["resolved_at_null"]; ok {
			query = query.Where("resolved_at IS NULL")
		}

		if _, ok := filters["resolved_at_not_null"]; ok {
			query = query.Where("resolved_at IS NOT NULL")
		}

		if from, ok := filters["from"].(time.Time); ok && !from.IsZero() {
			query = query.Where("detected_at >= ?", from)
		}

		if to, ok := filters["to"].(time.Time); ok && !to.IsZero() {
			query = query.Where("detected_at <= ?", to)
		}

		if includeRelations, ok := filters["include_relations"].(bool); ok && includeRelations {
			query = query.Preload("AlertType").Preload("Camera")
		}
	}

	// Get total count for pagination
	countQuery := query
	countQuery.Count(&total)

	// Get paginated results
	result := query.Limit(limit).Offset(offset).Find(&alerts)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return alerts, total, nil
}

// FindByID finds an alert by its ID
func (r *AlertRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Alert, error) {
	var alert entity.Alert

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&alert)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert not found")
		}
		return nil, result.Error
	}

	return &alert, nil
}

// FindActive retrieves currently active alerts with pagination
func (r *AlertRepositoryImpl) FindActive(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.Alert, int64, int64, error) {
	var alerts []entity.Alert
	var total int64
	var totalAllActive int64

	// Calculate offset
	offset := (page - 1) * limit

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	// Get total count of ALL active alerts (without any filters except active status)
	allActiveQuery := r.db.WithContext(ctx).Where("is_active = ?", true).
		Where("detected_at >= ? AND detected_at <= ?", startOfDay, endOfDay)
	allActiveQuery.Model(&entity.Alert{}).Count(&totalAllActive)

	// Build query for filtered active alerts
	query := r.db.WithContext(ctx).Model(&entity.Alert{}).Where("is_active = ?", true).
		Order("detected_at DESC")

	// Apply additional filters if provided
	if filters != nil {
		if alertTypeID, ok := filters["alert_type_id"].(uint); ok && alertTypeID != 0 {
			query = query.Where("alert_type_id = ?", alertTypeID)
		}

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
			query = query.Preload("AlertType").Preload("Camera")
		}
	}

	// Get total count for pagination (with filters applied)
	countQuery := query
	countQuery.Count(&total)

	// Get paginated results
	result := query.Limit(limit).Offset(offset).Find(&alerts)
	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return alerts, total, totalAllActive, nil
}

// Create adds a new alert to the database
func (r *AlertRepositoryImpl) Create(ctx context.Context, alert *entity.Alert) error {
	// Set detected_at to current time if not provided
	if alert.DetectedAt.IsZero() {
		alert.DetectedAt = time.Now()
	}

	// Set is_active to true by default
	alert.IsActive = true

	result := r.db.WithContext(ctx).Create(alert)
	if result.Error != nil {
		return result.Error
	}

	// Load related entities if needed
	if alert.ID != "" {
		r.db.WithContext(ctx).Preload("AlertType").Preload("Camera").First(alert, "id = ?", alert.ID)
	}

	return nil
}

// Update updates an existing alert in the database
func (r *AlertRepositoryImpl) Update(ctx context.Context, alert *entity.Alert) error {
	// Check if alert exists
	var existingAlert entity.Alert
	result := r.db.WithContext(ctx).Where("id = ?", alert.ID).First(&existingAlert)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("alert not found")
		}
		return result.Error
	}

	// Update the alert
	now := time.Now()
	alert.UpdatedAt = now

	updateResult := r.db.WithContext(ctx).Model(&existingAlert).Updates(alert)
	if updateResult.Error != nil {
		return updateResult.Error
	}

	// Load related entities if needed
	if alert.ID != "" {
		r.db.WithContext(ctx).Preload("AlertType").Preload("Camera").First(alert, "id = ?", alert.ID)
	}

	return nil
}

// Resolve marks an alert as resolved
func (r *AlertRepositoryImpl) Resolve(ctx context.Context, id string, resolvedBy, note string) error {
	// Check if alert exists
	var alert entity.Alert
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&alert)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("alert not found")
		}
		return result.Error
	}

	// Check if already resolved
	if !alert.IsActive {
		return errors.New("alert is already resolved")
	}

	// Update fields
	now := time.Now()

	updateResult := r.db.WithContext(ctx).Model(&alert).Updates(map[string]interface{}{
		"is_active":       false,
		"resolved_at":     now,
		"resolved_by":     resolvedBy,
		"resolution_note": note,
		"updated_at":      now,
	})

	return updateResult.Error
}
