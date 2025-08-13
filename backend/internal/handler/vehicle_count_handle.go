package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/service"
	"people-counting/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type VehicleCountData struct {
	UUID               string  `json:"uuid"`
	CctvID             uint    `json:"cctv_id"`
	InCountCar         int     `json:"in_count_car"`
	InCountTruck       int     `json:"in_count_truck"`
	InCountPeople      int     `json:"in_count_people"`
	OutCount           int     `json:"out_count"`
	DeviceTimestamp    string  `json:"device_timestamp"`
	DeviceTimestampUTC float64 `json:"device_timestamp_utc"`
}

// VehicleCountHandler handles HTTP requests related to vehicle counts
type VehicleCountHandler struct {
	vehicleCountService service.VehicleCountService
	cameraService       service.CameraService
}

// NewVehicleCountHandler creates a new vehicle count handler
func NewVehicleCountHandler(vehicleCountService service.VehicleCountService, cameraService service.CameraService) *VehicleCountHandler {
	return &VehicleCountHandler{
		vehicleCountService: vehicleCountService,
		cameraService:       cameraService,
	}
}

// RegisterRoutes registers routes for this handler
func (h *VehicleCountHandler) RegisterRoutes(router fiber.Router) {
	vehicles := router.Group("/vehicles")

	vehicles.Get("/", h.GetAllCounts)
	vehicles.Post("/", h.RecordCount)
	vehicles.Get("/summary", h.GetCountsSummary)
	vehicles.Get("/trends", h.GetCountsTrend)
	vehicles.Get("/distribution", h.GetCountsDistribution)
	vehicles.Get("/peak-hours", h.GetPeakHours)
	vehicles.Get("/latest/:cctv_id", h.GetLatestByCctv)
}

// GetAllCounts handles getting paginated vehicle count records
func (h *VehicleCountHandler) GetAllCounts(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	// Get filter parameters
	cctvID := c.Query("cctv_id", "")
	from := c.Query("from", "")
	to := c.Query("to", "")
	includeCctv := c.Query("include_cctv") == "true"

	counts, total, err := h.vehicleCountService.GetAllCounts(ctx, page, limit, cctvID, from, to, includeCctv)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if err.Error() == "invalid cctv ID" ||
			err.Error() == "invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" ||
			err.Error() == "invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "cctv not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"count": len(counts),
		"total": total,
		"page":  page,
		"pages": (total + int64(limit) - 1) / int64(limit),
		"data":  counts,
	})
}

// RecordCount handles adding a new vehicle count record
func (h *VehicleCountHandler) RecordCount(c *fiber.Ctx) error {
	ctx := c.Context()

	count := new(entity.VehicleCount)

	// Parse request body
	if err := c.BodyParser(count); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Record count
	err := h.vehicleCountService.RecordCount(ctx, count)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for validation errors
		if err.Error() == "cctv ID is required" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "cctv not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   "Vehicle count recorded successfully",
		"data":  count,
	})
}

// GetCountsSummary handles getting the summary of current vehicle counts
func (h *VehicleCountHandler) GetCountsSummary(c *fiber.Ctx) error {
	ctx := c.Context()

	cctvID := c.Query("cctv_id", "") // optional cctv filter

	dateRange, err := utils.ParseDateRangeFromQuery(c)
	if err != nil {
		return utils.HandleDateFilterError(c, err)
	}

	// Convert to strings for service layer
	from, to := dateRange.ToRFC3339Strings()

	summary, err := h.vehicleCountService.GetCountsSummary(ctx, cctvID, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Error getting vehicle counts summary: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  summary,
	})
}

// GetCountsTrend handles getting trend data for vehicle counts
func (h *VehicleCountHandler) GetCountsTrend(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get parameters
	interval := c.Query("interval", "hour") // hour, day, week, month
	cctvID := c.Query("cctv_id", "")        // optional cctv filter

	dateRange, err := utils.ParseDateRangeFromQuery(c)
	if err != nil {
		return utils.HandleDateFilterError(c, err)
	}

	// Convert to strings for service layer
	from, to := dateRange.ToRFC3339Strings()

	trends, err := h.vehicleCountService.GetCountsTrend(ctx, interval, cctvID, from, to)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid interval. Must be hour, day, week, or month" ||
			err.Error() == "invalid cctv ID" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "cctv not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"count": len(trends.Data),
		"data":  trends,
	})
}

// GetCountsDistribution handles getting distribution data
func (h *VehicleCountHandler) GetCountsDistribution(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get distribution type from query parameter
	distType := c.Query("type", "cctv") // cctv, vehicle_type

	// Used to limit the time window, default to last 24 hours
	timeWindow := c.Query("window", "24h") // 24h, 7d, 30d, all

	distribution, err := h.vehicleCountService.GetCountsDistribution(ctx, distType, timeWindow)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid distribution type. Must be cctv or vehicle_type" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  distribution,
	})
}

// GetPeakHours handles getting peak hours analysis for vehicle counts
func (h *VehicleCountHandler) GetPeakHours(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get parameters
	cctvID := c.Query("cctv_id", "") // optional cctv filter
	days := c.QueryInt("days", 7)    // default to last 7 days

	peakHours, err := h.vehicleCountService.GetPeakHours(ctx, cctvID, days)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid cctv ID" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "cctv not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"count": len(peakHours),
		"data":  peakHours,
	})
}

// GetLatestByCctv handles getting the latest vehicle count for a specific CCTV
func (h *VehicleCountHandler) GetLatestByCctv(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get CCTV ID from path parameter
	cctvIDStr := c.Params("cctv_id")
	if cctvIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "CCTV ID is required",
		})
	}

	latest, err := h.vehicleCountService.GetLatestByCctv(ctx, cctvIDStr)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid cctv ID" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "no vehicle count record found for this CCTV" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  latest,
	})
}

func (h *VehicleCountHandler) GetName() string {
	return "vehicle_counting_handler"
}

func (h *VehicleCountHandler) ProcessFile(filePath string) error {
	// Create a background context
	ctx := context.Background()

	// Read file contents
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON into VehicleCountData struct
	var countingData VehicleCountData
	if err := json.Unmarshal(fileData, &countingData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if vehicle count data already exists
	existingRecord, err := h.vehicleCountService.GetVehicleCountByID(ctx, countingData.UUID)
	isUpdate := false
	if err == nil && existingRecord != nil {
		// Data already exists, will update instead of creating
		isUpdate = true
	}

	// Convert VehicleCountData to VehicleCount entity
	counting, err := countingData.ToModel()
	if err != nil {
		return fmt.Errorf("failed to convert data to vehicle count: %w", err)
	}

	// Check if camera exists and create if it doesn't
	camera, err := h.cameraService.GetCameraByID(ctx, counting.CctvID)
	if err != nil || camera == nil {
		newCamera := &entity.Camera{
			ID:     counting.CctvID,
			Name:   fmt.Sprintf("CCTV %d (Auto-created)", counting.CctvID),
			Status: "active",
		}

		err := h.cameraService.CreateCamera(ctx, newCamera)
		if err != nil {
			return fmt.Errorf("failed to create camera: %w", err)
		}
	}

	// Create or update vehicle count data in database
	if isUpdate {
		if err := h.vehicleCountService.UpdateVehicleCount(ctx, counting); err != nil {
			return fmt.Errorf("failed to update vehicle counting data: %w", err)
		}
	} else {
		if err := h.vehicleCountService.CreateVehicleCount(ctx, counting); err != nil {
			return fmt.Errorf("failed to save vehicle counting data: %w", err)
		}
	}
	return nil
}

func (v VehicleCountData) ToModel() (*entity.VehicleCount, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp, err := time.ParseInLocation("2006-01-02T15:04:05.999999", v.DeviceTimestamp, loc)
	if err != nil {
		// Coba format alternatif jika RFC3339Nano gagal
		timestamp, err = time.ParseInLocation("2006-01-02T15:04:05.999999", v.DeviceTimestamp, loc)
		if err != nil {
			return nil, err
		}
	}

	// Use camera helper to determine camera ID
	cameraResolver := utils.NewCameraIDResolver()
	cameraID := cameraResolver.ResolveCameraID(0, v.CctvID) // Only CctvID is available in this struct

	// Create new VehicleCount entity
	vehicleCount := &entity.VehicleCount{
		ID:                 v.UUID,
		CctvID:             cameraID,
		Timestamp:          timestamp,
		InCountCar:         v.InCountCar,
		InCountTruck:       v.InCountTruck,
		InCountPeople:      v.InCountPeople,
		OutCount:           v.OutCount,
		DeviceTimestamp:    timestamp,
		DeviceTimestampUTC: v.DeviceTimestampUTC,
	}

	// Calculate total counts
	vehicleCount.CalculateTotalCounts()

	return vehicleCount, nil
}
