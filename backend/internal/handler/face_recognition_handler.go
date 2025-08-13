package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/service"
	"people-counting/pkg/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RecognitionData holds the data structure for incoming face recognition data
type RecognitionData struct {
	UUID      string `json:"uuid"`
	CameraID  *uint  `json:"camera_id"`
	ObjectID  string `json:"objectID"`
	TimeStamp string `json:"time_stamp"`
	ImagePath string `json:"image_path"`
}

// FaceRecognitionHandler handles HTTP requests related to face recognitions
type FaceRecognitionHandler struct {
	faceRecognitionService service.FaceRecognitionService
	cameraService          service.CameraService
}

// NewFaceRecognitionHandler creates a new face recognition handler
func NewFaceRecognitionHandler(faceRecognitionService service.FaceRecognitionService, cameraService service.CameraService) *FaceRecognitionHandler {
	return &FaceRecognitionHandler{
		faceRecognitionService: faceRecognitionService,
		cameraService:          cameraService,
	}
}

// RegisterRoutes registers routes for this handler
func (h *FaceRecognitionHandler) RegisterRoutes(router fiber.Router) {
	recognition := router.Group("/recognitions")

	recognition.Get("/", h.ListAll)
}

// ListAll handles getting paginated face recognition records
func (h *FaceRecognitionHandler) ListAll(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	// Get filter parameters
	cameraID := c.Query("camera_id", "")
	from := c.Query("from", "")
	to := c.Query("to", "")
	includeRelations := c.Query("include_relations") == "true"

	recognitions, total, err := h.faceRecognitionService.GetAll(ctx, page, limit, cameraID, from, to, includeRelations)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if err.Error() == "invalid camera ID" ||
			err.Error() == "invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" ||
			err.Error() == "invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "camera not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get base URL from request
	baseURL := c.Protocol() + "://" + c.Hostname()

	// Create a response structure to include image URLs
	type RecognitionResponse struct {
		ID         string         `json:"ID"`
		CameraID   uint           `json:"camera_id"`
		ImagePath  string         `json:"image_path"`
		ImageURL   string         `json:"image_url"`
		ObjectName string         `json:"object_name"`
		DetectedAt time.Time      `json:"detected_at"`
		CreatedAt  time.Time      `json:"created_at"`
		UpdatedAt  time.Time      `json:"updated_at"`
		Camera     *entity.Camera `json:"camera,omitempty"`
	}

	// Transform the response data
	var responseData []RecognitionResponse
	for _, recognition := range recognitions {
		// Extract filename from the original path
		filename := filepath.Base(strings.ReplaceAll(recognition.ImagePath, "\\", "/"))

		resp := RecognitionResponse{
			ID:         recognition.ID,
			CameraID:   recognition.CameraID,
			ImagePath:  recognition.ImagePath,
			ImageURL:   fmt.Sprintf("%s/api/images/faces/%s", baseURL, filename),
			ObjectName: recognition.ObjectName,
			DetectedAt: recognition.DetectedAt,
			CreatedAt:  recognition.CreatedAt,
			UpdatedAt:  recognition.UpdatedAt,
			Camera:     recognition.Camera,
		}
		responseData = append(responseData, resp)
	}

	return c.JSON(fiber.Map{
		"error": false,
		"count": len(responseData),
		"total": total,
		"page":  page,
		"pages": (total + int64(limit) - 1) / int64(limit),
		"data":  responseData,
	})
}

// GetName returns the handler name
func (h *FaceRecognitionHandler) GetName() string {
	return "face_recognition_handler"
}

// ProcessFile processes a face recognition file from the given path
func (h *FaceRecognitionHandler) ProcessFile(filePath string) error {
	// Create a background context
	ctx := context.Background()

	// Read file contents
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON into RecognitionData struct
	var recognitionData RecognitionData
	if err := json.Unmarshal(fileData, &recognitionData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if UUID is provided
	if recognitionData.UUID == "" {
		return fmt.Errorf("recognition UUID is required")
	}

	// Check if recognition with this UUID already exists
	existingData, err := h.faceRecognitionService.GetByID(ctx, recognitionData.UUID)
	isUpdate := false
	if err == nil && existingData != nil {
		// Recognition already exists, we will update it
		isUpdate = true
	}

	// Convert RecognitionData to FaceRecognition entity
	recognition, err := recognitionData.ToModel()
	if err != nil {
		return fmt.Errorf("failed to convert data to face recognition: %w", err)
	}

	// Check if camera exists and create if it doesn't
	camera, err := h.cameraService.GetCameraByID(ctx, recognition.CameraID)
	if err != nil || camera == nil {
		newCamera := &entity.Camera{
			ID:     recognition.CameraID,
			Name:   fmt.Sprintf("Camera %d (Auto-created)", recognition.CameraID),
			Status: "active",
		}

		err := h.cameraService.CreateCamera(ctx, newCamera)
		if err != nil {
			return fmt.Errorf("failed to create camera: %w", err)
		}
	}

	// Create or update face recognition in database
	if isUpdate {
		if err := h.faceRecognitionService.Update(ctx, recognition); err != nil {
			return fmt.Errorf("failed to update face recognition: %w", err)
		}
	} else {
		if err := h.faceRecognitionService.Create(ctx, recognition); err != nil {
			return fmt.Errorf("failed to create face recognition: %w", err)
		}
	}

	return nil
}

// ToModel converts RecognitionData to FaceRecognition entity
func (a RecognitionData) ToModel() (*entity.FaceRecognition, error) {
	parsedTime, err := utils.ParseDeviceTime(a.TimeStamp)
	if err != nil {
		return nil, err
	}

	// Use camera helper to determine camera ID
	cameraResolver := utils.NewCameraIDResolver()
	var cameraID uint
	if a.CameraID != nil {
		cameraID = cameraResolver.ResolveCameraID(*a.CameraID, 0)
	} else {
		cameraID = cameraResolver.ResolveCameraID(0, 0) // Will default to 1
	}

	data := &entity.FaceRecognition{
		ID:         a.UUID,
		CameraID:   cameraID,
		ObjectName: a.ObjectID,
		DetectedAt: parsedTime,
		ImagePath:  a.ImagePath,
	}

	return data, nil
}
