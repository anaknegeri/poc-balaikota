package repository

import (
	"context"
	"time"

	"people-counting/internal/domain/entity"
)

// CameraRepository defines the interface for camera data operations
type CameraRepository interface {
	FindAll(ctx context.Context, filters map[string]interface{}) ([]entity.Camera, error)
	FindByID(ctx context.Context, id uint) (*entity.Camera, error)
	FindByArea(ctx context.Context, areaID uint) ([]entity.Camera, error)
	Create(ctx context.Context, camera *entity.Camera) error
	Update(ctx context.Context, camera *entity.Camera) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	Delete(ctx context.Context, id uint) error
}

// PeopleCountRepository defines the interface for people count data operations
type PeopleCountRepository interface {
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.PeopleCount, int64, error)
	FindByID(ctx context.Context, id string) (*entity.PeopleCount, error)
	FindByArea(ctx context.Context, areaID uint, from, to time.Time, limit int) ([]entity.PeopleCount, error)
	Create(ctx context.Context, count *entity.PeopleCount) error
	Update(ctx context.Context, count *entity.PeopleCount) error
	GetSummary(ctx context.Context, filters map[string]interface{}) (*entity.CountSummary, error)
	GetTrends(ctx context.Context, interval string, filters map[string]interface{}) (*entity.CountsByTimeResult, error)
	GetDistributionByCamera(ctx context.Context, timeWindow time.Time) ([]entity.CameraSummary, error)
	GetDistributionByGender(ctx context.Context, timeWindow time.Time) (*entity.TotalCounts, error)
	GetDistributionByAge(ctx context.Context, timeWindow time.Time) (*entity.TotalCounts, error)
	GetPeakHoursAnalysis(ctx context.Context, filters map[string]interface{}) (*entity.PeakHoursAnalysis, error)
}

type VehicleCountRepository interface {
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.VehicleCount, int64, error)
	FindByID(ctx context.Context, id string) (*entity.VehicleCount, error)
	FindByCctv(ctx context.Context, cctvID uint, from, to time.Time, limit int) ([]entity.VehicleCount, error)
	Create(ctx context.Context, count *entity.VehicleCount) error
	Update(ctx context.Context, count *entity.VehicleCount) error

	GetSummary(ctx context.Context, filters map[string]interface{}) (*entity.VehicleCountSummary, error)
	GetTrends(ctx context.Context, interval string, filters map[string]interface{}) (*entity.VehicleCountsByTimeResult, error)
	GetDistributionByCctv(ctx context.Context, timeWindow time.Time) ([]entity.CctvVehicleSummary, error)
	GetDistributionByVehicleType(ctx context.Context, timeWindow time.Time) (*entity.VehicleTotalCounts, error)

	GetLatestByCctv(ctx context.Context, cctvID uint) (*entity.VehicleCount, error)
	GetCountsByTimeRange(ctx context.Context, from, to time.Time, cctvID *uint) ([]entity.VehicleCount, error)
	GetPeakHours(ctx context.Context, cctvID *uint, days int) ([]entity.VehicleTrendPoint, error)
}

// AlertTypeRepository defines the interface for alert type data operations
type AlertTypeRepository interface {
	FindAll(ctx context.Context) ([]entity.AlertType, error)
	FindByID(ctx context.Context, id uint) (*entity.AlertType, error)
	Create(ctx context.Context, alertType *entity.AlertType) (*entity.AlertType, error)
	FindByName(ctx context.Context, typeName string) (*entity.AlertType, error)
}

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.Alert, int64, error)
	FindByID(ctx context.Context, id string) (*entity.Alert, error)
	FindActive(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.Alert, int64, int64, error)
	Create(ctx context.Context, alert *entity.Alert) error
	Update(ctx context.Context, alert *entity.Alert) error
	Resolve(ctx context.Context, id string, resolvedBy, note string) error
}

type FaceRecognitionRepository interface {
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.FaceRecognition, int64, error)
	FindByID(ctx context.Context, id string) (*entity.FaceRecognition, error)
	Create(ctx context.Context, alert *entity.FaceRecognition) error
	Update(ctx context.Context, alert *entity.FaceRecognition) error
}
