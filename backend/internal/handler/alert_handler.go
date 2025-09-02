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
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FlexibleUint is a custom type that can unmarshal both string and int as uint
type FlexibleUint uint

func (f *FlexibleUint) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as uint first
	var uintVal uint
	if err := json.Unmarshal(data, &uintVal); err == nil {
		*f = FlexibleUint(uintVal)
		return nil
	}

	// If that fails, try to unmarshal as string and convert
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err
	}

	// Convert string to uint
	if strVal == "" {
		*f = FlexibleUint(0)
		return nil
	}

	parsedVal, err := strconv.ParseUint(strVal, 10, 32)
	if err != nil {
		return fmt.Errorf("cannot convert string '%s' to uint: %v", strVal, err)
	}

	*f = FlexibleUint(parsedVal)
	return nil
}

// ToUint converts FlexibleUint to uint
func (f FlexibleUint) ToUint() uint {
	return uint(f)
}

// FlexibleMissingItem is a custom type that can unmarshal both string and array as string
type FlexibleMissingItem string

func (f *FlexibleMissingItem) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*f = FlexibleMissingItem(strVal)
		return nil
	}

	// If that fails, try to unmarshal as array and join elements
	var arrVal []string
	if err := json.Unmarshal(data, &arrVal); err != nil {
		return err
	}

	// Join array elements with comma separator
	if len(arrVal) == 0 {
		*f = FlexibleMissingItem("")
	} else {
		*f = FlexibleMissingItem(strings.Join(arrVal, ", "))
	}
	return nil
}

// ToString converts FlexibleMissingItem to string
func (f FlexibleMissingItem) ToString() string {
	return string(f)
}

type AlertData struct {
	UUID         string              `json:"uuid"`
	CCTVID       uint                `json:"cctv_id"`
	CameraID     uint                `json:"camera_id"` // Alternative field name
	ObjectID     FlexibleUint        `json:"objectID"`
	ObjectName   string              `json:"object_name"` // Object type detected
	TimeStamp    string              `json:"time_stamp"`
	AlertType    string              `json:"alert_type"`
	ImagePath    string              `json:"image_path"`   // Standard field name
	ImagePathAlt string              `json:"Image_path"`   // Alternative with capital I
	MissingItem  FlexibleMissingItem `json:"missing_item"` // For PPE alerts - can be string or array
}

// AlertResponse represents the alert response with image URL
type AlertResponse struct {
	ID             string            `json:"id"`
	AlertTypeID    uint              `json:"alert_type_id"`
	CameraID       uint              `json:"camera_id"`
	Message        string            `json:"message"`
	Severity       string            `json:"severity"`
	IsActive       bool              `json:"is_active"`
	DetectedAt     time.Time         `json:"detected_at"`
	ResolvedAt     *time.Time        `json:"resolved_at"`
	ResolvedBy     string            `json:"resolved_by"`
	ResolutionNote string            `json:"resolution_note"`
	ImagePath      string            `json:"image_path"`
	ImageURL       string            `json:"image_url"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	AlertType      *entity.AlertType `json:"alert_type,omitempty"`
	Camera         *entity.Camera    `json:"camera,omitempty"`
}

// AlertHandler handles HTTP requests related to alerts
type AlertHandler struct {
	alertService     service.AlertService
	cameraService    service.CameraService
	alertTypeService service.AlertTypeService
	webSocketService service.WebSocketService
	baseFolder       string // Base folder for alert data
	alertType        string // Specific alert type for this handler
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertTypeService service.AlertTypeService, alertService service.AlertService, cameraService service.CameraService, webSocketService service.WebSocketService) *AlertHandler {
	return &AlertHandler{
		alertService:     alertService,
		cameraService:    cameraService,
		alertTypeService: alertTypeService,
		webSocketService: webSocketService,
		baseFolder:       "", // Will be set when used with specific folder
	}
}

// NewAlertHandlerWithFolder creates a new alert handler with specific base folder
func NewAlertHandlerWithFolder(alertTypeService service.AlertTypeService, alertService service.AlertService, cameraService service.CameraService, webSocketService service.WebSocketService, baseFolder string) *AlertHandler {
	return &AlertHandler{
		alertService:     alertService,
		cameraService:    cameraService,
		alertTypeService: alertTypeService,
		webSocketService: webSocketService,
		baseFolder:       baseFolder,
		alertType:        "", // Will be determined from folder structure or file content
	}
}

// NewAlertHandlerWithType creates a new alert handler with specific alert type
func NewAlertHandlerWithType(alertTypeService service.AlertTypeService, alertService service.AlertService, cameraService service.CameraService, webSocketService service.WebSocketService, baseFolder string, alertType string) *AlertHandler {
	return &AlertHandler{
		alertService:     alertService,
		cameraService:    cameraService,
		alertTypeService: alertTypeService,
		webSocketService: webSocketService,
		baseFolder:       baseFolder,
		alertType:        alertType,
	}
}

// RegisterRoutes registers routes for this handler
func (h *AlertHandler) RegisterRoutes(router fiber.Router) {
	alerts := router.Group("/alerts")

	alerts.Get("/", h.ListAlerts)
	alerts.Get("/active", h.GetActiveAlerts)
	alerts.Get("/types", h.GetAlertTypes)
	alerts.Post("/", h.CreateAlert)
	alerts.Put("/:id/resolve", h.ResolveAlert)
}

// ListAlerts handles getting paginated alert records
func (h *AlertHandler) ListAlerts(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	// Get filter parameters
	isActive := c.Query("is_active", "")
	alertTypeID := c.Query("alert_type_id", "")
	cameraID := c.Query("camera_id", "")
	from := c.Query("from", "")
	to := c.Query("to", "")
	search := c.Query("search", "")
	severity := c.Query("severity", "")
	status := c.Query("status", "")
	includeRelations := c.Query("include_relations") == "true"

	// Set default date range to current month if no dates provided
	if from == "" && to == "" {
		now := time.Now()
		// First day of current month
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		// Last day of current month
		lastDay := firstDay.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		from = firstDay.Format(time.RFC3339)
		to = lastDay.Format(time.RFC3339)
	}

	alerts, total, err := h.alertService.GetAllAlerts(ctx, page, limit, isActive, alertTypeID, cameraID, from, to, search, severity, status, includeRelations)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if err.Error() == "invalid alert type ID" ||
			err.Error() == "invalid camera ID" ||
			err.Error() == "invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" ||
			err.Error() == "invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "alert type not found" ||
			err.Error() == "camera not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get base URL from request
	baseURL := c.Protocol() + "://" + c.Hostname()

	// Transform the response data to include image URLs
	var responseData []AlertResponse
	for _, alert := range alerts {
		resp := AlertResponse{
			ID:             alert.ID,
			AlertTypeID:    alert.AlertTypeID,
			CameraID:       alert.CameraID,
			Message:        alert.Message,
			Severity:       alert.Severity,
			IsActive:       alert.IsActive,
			DetectedAt:     alert.DetectedAt,
			ResolvedAt:     alert.ResolvedAt,
			ResolvedBy:     alert.ResolvedBy,
			ResolutionNote: alert.ResolutionNote,
			ImagePath:      alert.ImageURL, // Store original path
			CreatedAt:      alert.CreatedAt,
			UpdatedAt:      alert.UpdatedAt,
			AlertType:      &alert.AlertType,
			Camera:         alert.Camera,
		}

		// Generate image URL if image path exists
		if alert.ImageURL != "" {
			// Extract filename from the original path
			filename := filepath.Base(strings.ReplaceAll(alert.ImageURL, "\\", "/"))

			// Use alert type name for URL path if available
			if alert.AlertType.Name != "" {
				alertTypeName := strings.ToLower(strings.ReplaceAll(alert.AlertType.Name, " ", "-"))
				imageURL := fmt.Sprintf("%s/api/images/alerts/%s/%s", baseURL, alertTypeName, filename)
				resp.ImageURL = imageURL
			} else {
				// Fallback to general alerts path
				imageURL := fmt.Sprintf("%s/api/images/alerts/%s", baseURL, filename)
				resp.ImageURL = imageURL
			}
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

// Updated GetActiveAlerts handler using the helper
func (h *AlertHandler) GetActiveAlerts(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	// Get filter parameters
	alertTypeID := c.Query("alert_type_id", "")
	cameraID := c.Query("camera_id", "")
	includeRelations := c.Query("include_relations") == "true"

	// Parse date range using helper
	dateRange, err := utils.ParseDateRangeFromQuery(c)
	if err != nil {
		return utils.HandleDateFilterError(c, err)
	}

	// Convert to strings for service layer
	from, to := dateRange.ToRFC3339Strings()

	alerts, total, totalAllActive, err := h.alertService.GetActiveAlerts(ctx, page, limit, alertTypeID, cameraID, from, to, includeRelations)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if strings.Contains(err.Error(), "invalid alert type ID") ||
			strings.Contains(err.Error(), "invalid camera ID") {
			status = fiber.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get base URL from request
	baseURL := c.Protocol() + "://" + c.Hostname()

	// Transform the response data to include image URLs
	var responseData []AlertResponse
	for _, alert := range alerts {
		resp := AlertResponse{
			ID:             alert.ID,
			AlertTypeID:    alert.AlertTypeID,
			CameraID:       alert.CameraID,
			Message:        alert.Message,
			Severity:       alert.Severity,
			IsActive:       alert.IsActive,
			DetectedAt:     alert.DetectedAt,
			ResolvedAt:     alert.ResolvedAt,
			ResolvedBy:     alert.ResolvedBy,
			ResolutionNote: alert.ResolutionNote,
			ImagePath:      alert.ImageURL,
			CreatedAt:      alert.CreatedAt,
			UpdatedAt:      alert.UpdatedAt,
			AlertType:      &alert.AlertType,
			Camera:         alert.Camera,
		}

		// Generate image URL if image path exists
		if alert.ImageURL != "" {
			filename := filepath.Base(strings.ReplaceAll(alert.ImageURL, "\\", "/"))

			// Use alert type name for URL path if available
			if alert.AlertType.Name != "" {
				alertTypeName := strings.ToLower(strings.ReplaceAll(alert.AlertType.Name, " ", "-"))
				imageURL := fmt.Sprintf("%s/api/images/alerts/%s/%s", baseURL, alertTypeName, filename)
				resp.ImageURL = imageURL
			} else {
				// Fallback to general alerts path
				imageURL := fmt.Sprintf("%s/api/images/alerts/%s", baseURL, filename)
				resp.ImageURL = imageURL
			}
		}

		responseData = append(responseData, resp)
	}

	return c.JSON(fiber.Map{
		"error":            false,
		"count":            len(responseData),
		"total":            total,
		"count_all_active": totalAllActive,
		"page":             page,
		"pages":            (total + int64(limit) - 1) / int64(limit),
		"data":             responseData,
	})
}

// CreateAlert handles creating a new alert
func (h *AlertHandler) CreateAlert(c *fiber.Ctx) error {
	ctx := c.Context()

	alert := new(entity.Alert)

	// Parse request body
	if err := c.BodyParser(alert); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Create alert
	err := h.alertService.CreateAlert(ctx, alert)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for validation errors
		if err.Error() == "alert type ID is required" ||
			err.Error() == "message is required" ||
			err.Error() == "camera ID is required" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "alert type not found" ||
			err.Error() == "camera not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Transform response to include image URL
	baseURL := c.Protocol() + "://" + c.Hostname()
	resp := AlertResponse{
		ID:             alert.ID,
		AlertTypeID:    alert.AlertTypeID,
		CameraID:       alert.CameraID,
		Message:        alert.Message,
		Severity:       alert.Severity,
		IsActive:       alert.IsActive,
		DetectedAt:     alert.DetectedAt,
		ResolvedAt:     alert.ResolvedAt,
		ResolvedBy:     alert.ResolvedBy,
		ResolutionNote: alert.ResolutionNote,
		ImagePath:      alert.ImageURL,
		CreatedAt:      alert.CreatedAt,
		UpdatedAt:      alert.UpdatedAt,
		AlertType:      &alert.AlertType,
		Camera:         alert.Camera,
	}

	// Generate image URL if image path exists
	if alert.ImageURL != "" {
		filename := filepath.Base(strings.ReplaceAll(alert.ImageURL, "\\", "/"))

		// Use alert type name for URL path if available
		if alert.AlertType.Name != "" {
			alertTypeName := strings.ToLower(strings.ReplaceAll(alert.AlertType.Name, " ", "-"))
			imageURL := fmt.Sprintf("%s/api/images/alerts/%s/%s", baseURL, alertTypeName, filename)
			resp.ImageURL = imageURL
		} else {
			// Fallback to general alerts path
			imageURL := fmt.Sprintf("%s/api/images/alerts/%s", baseURL, filename)
			resp.ImageURL = imageURL
		}
	}

	// Send WebSocket notification to all connected clients
	if h.webSocketService != nil {
		cameraName := ""
		if alert.Camera != nil {
			cameraName = alert.Camera.Name
		}
		alertTypeName := ""
		if alert.AlertType.Name != "" {
			alertTypeName = alert.AlertType.Name
		}
		h.webSocketService.NotifyAlert(alertTypeName, cameraName, alert.Message, resp)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   "Alert created successfully",
		"data":  resp,
	})
}

// ResolveAlert handles resolving (deactivating) an alert
func (h *AlertHandler) ResolveAlert(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get alert ID from path parameter
	id := c.Params("id")

	// Parse request body
	type ResolutionData struct {
		ResolvedBy     string `json:"resolved_by"`
		ResolutionNote string `json:"resolution_note"`
	}

	resData := new(ResolutionData)
	if err := c.BodyParser(resData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Resolve alert
	err := h.alertService.ResolveAlert(ctx, id, resData.ResolvedBy, resData.ResolutionNote)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if err.Error() == "alert ID is required" ||
			err.Error() == "resolved by is required" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "alert not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "alert is already resolved" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   "Alert resolved successfully",
	})
}

func (h *AlertHandler) GetName() string {
	return "alert_handler"
}

func (h *AlertHandler) ProcessFile(filePath string) error {
	return h.ProcessFileWithType(filePath, h.alertType)
}

func (h *AlertHandler) ProcessFileWithType(filePath string, alertType string) error {
	// Create a background context
	ctx := context.Background()

	// Read file contents
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON into AlertData struct
	var alertData AlertData
	if err := json.Unmarshal(fileData, &alertData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if UUID is provided
	if alertData.UUID == "" {
		return fmt.Errorf("alert UUID is required")
	}

	// Check if alert with this UUID already exists
	existingAlert, err := h.alertService.GetAlertByID(ctx, alertData.UUID)
	isUpdate := false
	if err == nil && existingAlert != nil {
		// Alert already exists, we will update it
		isUpdate = true
	}

	// Use provided alert type if available, otherwise use alert type from JSON data
	finalAlertType := alertData.AlertType
	if alertType != "" {
		finalAlertType = alertType
		alertData.AlertType = alertType // Update the struct as well
	}

	// Set alert type based on alert type string
	alertTypeID, err := h.getAlertTypeIDFromString(ctx, strings.TrimSpace(finalAlertType))
	if err != nil {
		return fmt.Errorf("failed to get alert type ID: %w", err)
	}

	// Adjust image path based on alert type and base folder
	// Handle both image_path and Image_path fields
	originalImagePath := alertData.ImagePath
	if originalImagePath == "" && alertData.ImagePathAlt != "" {
		originalImagePath = alertData.ImagePathAlt
	}

	if h.baseFolder != "" && originalImagePath != "" {
		// Extract filename from original path
		imageName := filepath.Base(originalImagePath)

		// Determine alert type for URL path
		alertTypeName := strings.ToLower(strings.TrimSpace(finalAlertType))

		// Create new path using alert type specific URL: /api/images/alerts/{type}/{filename}
		processedImagePath := fmt.Sprintf("/api/images/alerts/%s/%s", alertTypeName, imageName)

		// Update both fields to ensure consistency
		alertData.ImagePath = processedImagePath
		alertData.ImagePathAlt = processedImagePath
	}

	// Convert AlertData to Alert entity
	alert, err := alertData.ToAlert()
	if err != nil {
		return fmt.Errorf("failed to convert data to alert: %w", err)
	}

	// Check if camera exists and create if it doesn't
	if alert.CameraID != 0 {
		camera, err := h.cameraService.GetCameraByID(ctx, alert.CameraID)

		// Only create camera if it doesn't exist (when we get an error OR camera is nil)
		if err != nil || camera == nil {
			fmt.Printf("Camera %d not found, creating new camera\n", alert.CameraID)

			newCamera := &entity.Camera{
				ID:     alert.CameraID,
				Name:   fmt.Sprintf("Camera %d (Auto-created)", alert.CameraID),
				Status: "active",
			}

			err := h.cameraService.CreateCamera(ctx, newCamera)
			if err != nil {
				// Check if error is due to camera already existing (race condition)
				if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
					strings.Contains(strings.ToLower(err.Error()), "already exists") ||
					strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
					fmt.Printf("Camera %d already exists (race condition), continuing...\n", alert.CameraID)
				} else {
					return fmt.Errorf("failed to create camera: %w", err)
				}
			} else {
				fmt.Printf("Successfully created camera %d\n", alert.CameraID)
			}
		}
	}

	alert.AlertTypeID = alertTypeID

	// Create or update alert in database
	if isUpdate {
		if err := h.alertService.UpdateAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to update alert: %w", err)
		}
	} else {
		if err := h.alertService.CreateAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}
	}

	// Send WebSocket notification after successful alert creation/update
	if h.webSocketService != nil {
		// Get camera name for notification
		cameraName := ""
		if alert.Camera != nil {
			cameraName = alert.Camera.Name
		} else {
			cameraName = fmt.Sprintf("Camera %d", alert.CameraID)
		}

		// Get alert type name for notification
		alertTypeName := ""
		if alert.AlertType.Name != "" {
			alertTypeName = alert.AlertType.Name
		} else {
			alertTypeName = finalAlertType
		}

		// Create response data for WebSocket notification
		alertResponse := AlertResponse{
			ID:             alert.ID,
			AlertTypeID:    alert.AlertTypeID,
			CameraID:       alert.CameraID,
			Message:        alert.Message,
			Severity:       alert.Severity,
			IsActive:       alert.IsActive,
			DetectedAt:     alert.DetectedAt,
			ResolvedAt:     alert.ResolvedAt,
			ResolvedBy:     alert.ResolvedBy,
			ResolutionNote: alert.ResolutionNote,
			ImagePath:      alert.ImageURL,
			ImageURL:       alert.ImageURL,
			CreatedAt:      alert.CreatedAt,
			UpdatedAt:      alert.UpdatedAt,
			AlertType:      &alert.AlertType,
			Camera:         alert.Camera,
		}

		// Send WebSocket notification
		h.webSocketService.NotifyAlert(alertTypeName, cameraName, alert.Message, alertResponse)
	}

	return nil
}

func (a AlertData) ToAlert() (*entity.Alert, error) {
	// Use flexible date parser for better timestamp handling
	parser := utils.NewFlexibleDateParser()

	// Add alert-specific timestamp formats
	parser.AddCustomFormat("20060102_150405")        // YYYYMMDD_HHMMSS
	parser.AddCustomFormat("20060102_150405_000000") // YYYYMMDD_HHMMSS_microseconds

	// Parse the timestamp using flexible parser
	parsedTime, err := parser.Parse(a.TimeStamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp '%s': %v", a.TimeStamp, err)
	}

	// Use camera helper to determine camera ID (ObjectID should not be used for camera resolution)
	cameraResolver := utils.NewCameraIDResolver()
	cameraID := cameraResolver.ResolveCameraID(a.CameraID, a.CCTVID) // Pass 0 for objectID to exclude it

	// Determine which image path to use (handle both lowercase and uppercase variants)
	imagePath := a.ImagePath
	if imagePath == "" && a.ImagePathAlt != "" {
		imagePath = a.ImagePathAlt
	}

	alert := &entity.Alert{
		ID:         a.UUID,
		CameraID:   cameraID,
		Message:    a.getAlertMessage(),
		DetectedAt: parsedTime,
		ImageURL:   imagePath,
		Severity:   "high",
	}

	return alert, nil
}

func (h *AlertHandler) getAlertTypeIDFromString(ctx context.Context, typeName string) (uint, error) {
	// Try to find existing alert type in database
	alertType, err := h.alertTypeService.GetAlertTypeByName(ctx, typeName)
	if err == nil && alertType != nil {
		return alertType.ID, nil
	}

	// If not found, create new alert type
	newAlertType := &entity.AlertType{
		Name:        typeName,
		Description: fmt.Sprintf("Auto-generated alert type for %s", typeName),
		Icon:        "alert-triangle",
		Color:       "#ff6b35",
	}

	createdAlertType, err := h.alertTypeService.CreateAlertType(ctx, newAlertType)
	if err != nil {
		// If creation fails, try to get existing one again (in case of race condition)
		existingAlertType, getErr := h.alertTypeService.GetAlertTypeByName(ctx, typeName)
		if getErr == nil && existingAlertType != nil {
			return existingAlertType.ID, nil
		}
		return 0, fmt.Errorf("failed to create alert type: %w", err)
	}

	return createdAlertType.ID, nil
}

func (a *AlertData) getAlertMessage() string {
	alertType := strings.ToLower(strings.TrimSpace(a.AlertType))
	missingItem := a.MissingItem.ToString()

	// Handle Indonesian alert types
	switch alertType {
	case "tidak patuh":
		if missingItem != "" {
			return fmt.Sprintf("Personal protective equipment violation: Missing %s", missingItem)
		}
		return "Personal protective equipment violation detected"
	case "restricted":
		return "Person detected in restricted area"
	case "fall-detection":
		return "Person fall detected"
	case "loitering":
		return "Extended loitering detected"
	case "hazardous-area":
		return "Person detected in hazardous area"
	case "personal-protective-equipment":
		if missingItem != "" {
			return fmt.Sprintf("Personal protective equipment violation: Missing %s", missingItem)
		}
		return "Personal protective equipment violation detected"
	}

	// Default message for unknown alert types
	if missingItem != "" {
		return fmt.Sprintf("Alert detected: %s (Missing: %s)", a.AlertType, missingItem)
	}
	return fmt.Sprintf("Alert detected: %s", a.AlertType)
}

// GetAlertTypes handles getting all alert types
func (h *AlertHandler) GetAlertTypes(c *fiber.Ctx) error {
	ctx := c.Context()

	alertTypes, err := h.alertTypeService.GetAllAlertTypes(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Failed to retrieve alert types",
			"data":  nil,
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   "Alert types retrieved successfully",
		"data":  alertTypes,
		"count": len(alertTypes),
	})
}
