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

type PeopleCountData struct {
	UUID               string  `json:"uuid"`
	CCTVID             uint    `json:"cctv_id"`
	InCount            int     `json:"in_count"`
	OutCount           int     `json:"out_count"`
	FemaleCount        int     `json:"female_count"`
	MaleCount          int     `json:"male_count"`
	ChildrenCount      int     `json:"children_count"`
	AdultCount         int     `json:"adult_count"`
	ElderCount         int     `json:"elder_count"`
	DeviceTimestamp    string  `json:"device_timestamp"`
	DeviceTimestampUTC float64 `json:"device_timestamp_utc"`
	SyncStatus         bool    `json:"sync_status"`
}

// PeopleCountHandler handles HTTP requests related to people counts
type PeopleCountHandler struct {
	peopleCountService service.PeopleCountService
	cameraService      service.CameraService
}

// NewPeopleCountHandler creates a new people count handler
func NewPeopleCountHandler(peopleCountService service.PeopleCountService, cameraService service.CameraService) *PeopleCountHandler {
	return &PeopleCountHandler{
		peopleCountService: peopleCountService,
		cameraService:      cameraService,
	}
}

// RegisterRoutes registers routes for this handler
func (h *PeopleCountHandler) RegisterRoutes(router fiber.Router) {
	counts := router.Group("/counts")

	counts.Get("/", h.GetAllCounts)
	counts.Post("/", h.RecordCount)
	counts.Get("/summary", h.GetCountsSummary)
	counts.Get("/trends", h.GetCountsTrend)
	counts.Get("/distribution", h.GetCountsDistribution)
	counts.Get("/peak-hours", h.GetPeakHoursAnalysis)
}

// GetAllCounts handles getting paginated people count records
func (h *PeopleCountHandler) GetAllCounts(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	// Get filter parameters
	areaID := c.Query("area_id", "")
	from := c.Query("from", "")
	to := c.Query("to", "")
	includeArea := c.Query("include_area") == "true"

	counts, total, err := h.peopleCountService.GetAllCounts(ctx, page, limit, areaID, from, to, includeArea)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if err.Error() == "invalid area ID" ||
			err.Error() == "invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" ||
			err.Error() == "invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "area not found" {
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

// RecordCount handles adding a new people count record
func (h *PeopleCountHandler) RecordCount(c *fiber.Ctx) error {
	ctx := c.Context()

	count := new(entity.PeopleCount)

	// Parse request body
	if err := c.BodyParser(count); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Record count
	err := h.peopleCountService.RecordCount(ctx, count)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for validation errors
		if err.Error() == "area ID is required" ||
			err.Error() == "demographic counts (child + adult + elderly) must equal gender counts (male + female)" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "area not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   "Count recorded successfully",
		"data":  count,
	})
}

// GetCountsSummary handles getting the summary of current people counts
func (h *PeopleCountHandler) GetCountsSummary(c *fiber.Ctx) error {
	ctx := c.Context()

	dateRange, err := utils.ParseDateRangeFromQuery(c)
	if err != nil {
		return utils.HandleDateFilterError(c, err)
	}

	// Convert to strings for service layer
	from, to := dateRange.ToRFC3339Strings()

	summary, err := h.peopleCountService.GetCountsSummary(ctx, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Error getting counts summary: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  summary,
	})
}

// GetCountsTrend handles getting trend data for people counts
func (h *PeopleCountHandler) GetCountsTrend(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get parameters
	interval := c.Query("interval", "hour") // hour, day, week, month
	areaID := c.Query("area_id", "")        // optional area filter

	dateRange, err := utils.ParseDateRangeFromQuery(c)
	if err != nil {
		return utils.HandleDateFilterError(c, err)
	}

	// Convert to strings for service layer
	from, to := dateRange.ToRFC3339Strings()

	trends, err := h.peopleCountService.GetCountsTrend(ctx, interval, areaID, from, to)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid interval. Must be hour, day, week, or month" ||
			err.Error() == "invalid area ID" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "area not found" {
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
func (h *PeopleCountHandler) GetCountsDistribution(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get distribution type from query parameter
	distType := c.Query("type", "camera") // area, gender, age

	// Used to limit the time window, default to last 24 hours
	timeWindow := c.Query("window", "24h") // 24h, 7d, 30d, all

	distribution, err := h.peopleCountService.GetCountsDistribution(ctx, distType, timeWindow)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid distribution type. Must be camera, gender, or age" {
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

func (h *PeopleCountHandler) GetName() string {
	return "people_counting_handler"
}

func (h *PeopleCountHandler) ProcessFile(filePath string) error {
	// Create a background context
	ctx := context.Background()

	// Read file contents
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON into AlertData struct
	var countingData PeopleCountData
	if err := json.Unmarshal(fileData, &countingData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if people count data already exists
	existingCount, err := h.peopleCountService.GetByID(ctx, countingData.UUID)
	isUpdate := false
	if err == nil && existingCount != nil {
		// Data already exists, will update instead of creating
		isUpdate = true
	}

	// Convert AlertData to Alert entity
	counting, err := countingData.ToModel()
	if err != nil {
		return fmt.Errorf("failed to convert data to alert: %w", err)
	}

	// Check if camera exists and create if it doesn't
	camera, err := h.cameraService.GetCameraByID(ctx, counting.CameraID)
	if err != nil || camera == nil {
		newCamera := &entity.Camera{
			ID:     counting.CameraID,
			Name:   fmt.Sprintf("Camera %d (Auto-created)", counting.CameraID),
			Status: "active",
		}

		err := h.cameraService.CreateCamera(ctx, newCamera)
		if err != nil {
			return fmt.Errorf("failed to create camera: %w", err)
		}
	}

	// Create or update people count data in database
	if isUpdate {
		if err := h.peopleCountService.UpdatePeopleCount(ctx, counting); err != nil {
			return fmt.Errorf("failed to update counting data: %w", err)
		}
	} else {
		if err := h.peopleCountService.CreatePeopleCount(ctx, counting); err != nil {
			return fmt.Errorf("failed to save counting data: %w", err)
		}
	}
	return nil
}

func (p PeopleCountData) ToModel() (*entity.PeopleCount, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp, err := time.ParseInLocation("2006-01-02T15:04:05.999999", p.DeviceTimestamp, loc)
	if err != nil {
		// Coba format alternatif jika RFC3339Nano gagal
		timestamp, err = time.ParseInLocation("2006-01-02T15:04:05.999999", p.DeviceTimestamp, loc)
		if err != nil {
			return nil, err
		}
	}

	totalCount := p.InCount - p.OutCount

	// Use camera helper to determine camera ID
	cameraResolver := utils.NewCameraIDResolver()
	cameraID := cameraResolver.ResolveCameraID(0, p.CCTVID) // Only CCTVID is available in this struct

	// Buat entity PeopleCount baru
	peopleCount := &entity.PeopleCount{
		ID:           p.UUID,
		CameraID:     cameraID,
		Timestamp:    timestamp,
		MaleCount:    p.MaleCount,
		FemaleCount:  p.FemaleCount,
		ChildCount:   p.ChildrenCount,
		AdultCount:   p.AdultCount,
		ElderlyCount: p.ElderCount,
		TotalCount:   totalCount,
	}

	return peopleCount, nil
}

// GetPeakHoursAnalysis handles getting peak hours analysis
func (h *PeopleCountHandler) GetPeakHoursAnalysis(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get parameters
	cameraID := c.Query("camera_id", "") // optional camera filter
	dateRange, err := utils.ParseDateRangeFromQuery(c)
	if err != nil {
		return utils.HandleDateFilterError(c, err)
	}

	// Convert to strings for service layer
	from, to := dateRange.ToRFC3339Strings()

	analysis, err := h.peopleCountService.GetPeakHoursAnalysis(ctx, cameraID, from, to)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "invalid camera ID" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  analysis,
	})
}
