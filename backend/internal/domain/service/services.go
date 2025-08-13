package service

import (
	"context"
	"encoding/json"
	"time"

	"people-counting/internal/domain/entity"

	"github.com/gofiber/fiber/v2"
)

// CameraService defines the interface for camera business logic
type CameraService interface {
	GetAllCameras(ctx context.Context, status string) ([]entity.Camera, error)
	GetCameraByID(ctx context.Context, id uint) (*entity.Camera, error)
	CreateCamera(ctx context.Context, camera *entity.Camera) error
	UpdateCamera(ctx context.Context, camera *entity.Camera) error
	UpdateCameraStatus(ctx context.Context, id uint, status string) error
	DeleteCamera(ctx context.Context, id uint) error

	GetCameraStreamURL(c *fiber.Ctx, cameraID uint) string
	GetCameraImageURL(c *fiber.Ctx, cameraID uint) string
}

// PeopleCountService defines the interface for people count business logic
type PeopleCountService interface {
	GetAllCounts(ctx context.Context, page, limit int, areaID, from, to string, includeArea bool) ([]entity.PeopleCount, int64, error)
	RecordCount(ctx context.Context, count *entity.PeopleCount) error
	GetCountsSummary(ctx context.Context, from, to string) (*entity.CountSummary, error)
	GetCountsTrend(ctx context.Context, interval string, areaID, from, to string) (*entity.CountsByTimeResult, error)
	GetCountsDistribution(ctx context.Context, distType, timeWindow string) (interface{}, error)
	CreatePeopleCount(ctx context.Context, counting *entity.PeopleCount) error
	UpdatePeopleCount(ctx context.Context, counting *entity.PeopleCount) error
	GetByID(ctx context.Context, id string) (*entity.PeopleCount, error)
	GetAlertByID(ctx context.Context, id string) (*entity.PeopleCount, error)
	GetPeakHoursAnalysis(ctx context.Context, cameraID string, from, to string) (*entity.PeakHoursAnalysis, error)
}

// VehicleCountService defines the interface for vehicle count service operations
type VehicleCountService interface {
	// Basic operations
	GetVehicleCountByID(ctx context.Context, id string) (*entity.VehicleCount, error)
	GetAllCounts(ctx context.Context, page, limit int, cctvID, from, to string, includeCctv bool) ([]entity.VehicleCount, int64, error)
	CreateVehicleCount(ctx context.Context, count *entity.VehicleCount) error
	UpdateVehicleCount(ctx context.Context, count *entity.VehicleCount) error
	RecordCount(ctx context.Context, count *entity.VehicleCount) error

	// Summary and analytics
	GetCountsSummary(ctx context.Context, cctvID, from, to string) (*entity.VehicleCountSummary, error)
	GetCountsTrend(ctx context.Context, interval string, cctvID, from, to string) (*entity.VehicleCountsByTimeResult, error)
	GetCountsDistribution(ctx context.Context, distType, timeWindow string) (interface{}, error)

	// Specialized analytics
	GetPeakHours(ctx context.Context, cctvID string, days int) ([]entity.VehicleTrendPoint, error)
	GetLatestByCctv(ctx context.Context, cctvIDStr string) (*entity.VehicleCount, error)
	GetCountsByTimeRange(ctx context.Context, from, to time.Time, cctvID string) ([]entity.VehicleCount, error)

	// Advanced analytics
	CalculateOccupancyRate(ctx context.Context, cctvID string, from, to time.Time) (float64, error)
	GetBusiestCctvs(ctx context.Context, timeWindow time.Time, limit int) ([]entity.CctvVehicleSummary, error)
}

// AlertTypeService defines the interface for alert type business logic
type AlertTypeService interface {
	GetAllAlertTypes(ctx context.Context) ([]entity.AlertType, error)
	GetAlertTypeByName(ctx context.Context, typeName string) (*entity.AlertType, error)
	CreateAlertType(ctx context.Context, alertType *entity.AlertType) (*entity.AlertType, error)
}

// AlertService defines the interface for alert business logic
type AlertService interface {
	GetAllAlerts(ctx context.Context, page, limit int, isActive, alertTypeID, cameraID, from, to, search, severity, status string, includeRelations bool) ([]entity.Alert, int64, error)
	GetActiveAlerts(ctx context.Context, page, limit int, alertTypeID, cameraID, from, to string, includeRelations bool) ([]entity.Alert, int64, int64, error)
	CreateAlert(ctx context.Context, alert *entity.Alert) error
	UpdateAlert(ctx context.Context, alert *entity.Alert) error
	ResolveAlert(ctx context.Context, id, resolvedBy, resolutionNote string) error
	GetAlertByID(ctx context.Context, id string) (*entity.Alert, error)
}

// AnalyticsService defines the interface for analytics business logic
type AnalyticsService interface {
	GetOccupancyRate(ctx context.Context, areaID uint) (float64, error)
	GetPeakHours(ctx context.Context, areaID uint, date time.Time) ([]entity.TrendPoint, error)
	GetHistoricalComparison(ctx context.Context, areaID uint, period string) ([]entity.TrendPoint, error)
	GetDemographicInsights(ctx context.Context, areaID uint, timeWindow string) (interface{}, error)
}

// FaceRecognitionService defines the interface for analytics business logic
type FaceRecognitionService interface {
	Create(ctx context.Context, alert *entity.FaceRecognition) error
	Update(ctx context.Context, alert *entity.FaceRecognition) error
	GetByID(ctx context.Context, id string) (*entity.FaceRecognition, error)
	GetAll(ctx context.Context, page, limit int, isActive, from, to string, includeRelations bool) ([]entity.FaceRecognition, int64, error)
}

// WebSocketService defines the interface for WebSocket business logic
type WebSocketService interface {
	NotifyAlert(alertType, cameraName, message string, data interface{})
	SendPersonalizedMessage(clientID, messageType string, data interface{}) bool
	GetConnectionStats() map[string]interface{}
	HandleClientMessage(clientID string, messageType string, data json.RawMessage) error
}
