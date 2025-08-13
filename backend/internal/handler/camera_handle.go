package handler

import (
	"strconv"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/service"

	"github.com/gofiber/fiber/v2"
)

// CameraHandler handles HTTP requests related to cameras
type CameraHandler struct {
	cameraService service.CameraService
}

// NewCameraHandler creates a new camera handler
func NewCameraHandler(cameraService service.CameraService) *CameraHandler {
	return &CameraHandler{
		cameraService: cameraService,
	}
}

// RegisterRoutes registers routes for this handler
func (h *CameraHandler) RegisterRoutes(router fiber.Router) {
	cameras := router.Group("/cameras")

	cameras.Get("/", h.ListCameras)
	cameras.Get("/:id", h.GetCamera)
	cameras.Post("/", h.CreateCamera)
	cameras.Put("/:id", h.UpdateCamera)
	cameras.Delete("/:id", h.DeleteCamera)
	cameras.Put("/:id/status", h.UpdateCameraStatus)
}

// ListCameras handles getting all cameras
func (h *CameraHandler) ListCameras(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get query parameters
	status := c.Query("status")

	cameras, err := h.cameraService.GetAllCameras(ctx, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Error getting cameras: " + err.Error(),
		})
	}

	for i := range cameras {
		cameras[i].StreamURL = h.cameraService.GetCameraStreamURL(c, cameras[i].ID)
		cameras[i].ImageURL = h.cameraService.GetCameraImageURL(c, cameras[i].ID)
	}

	return c.JSON(fiber.Map{
		"error": false,
		"count": len(cameras),
		"data":  cameras,
	})
}

// GetCamera handles getting a single camera
func (h *CameraHandler) GetCamera(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse ID parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid camera ID",
		})
	}

	camera, err := h.cameraService.GetCameraByID(ctx, uint(id))
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "camera not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	camera.StreamURL = h.cameraService.GetCameraStreamURL(c, camera.ID)
	camera.ImageURL = h.cameraService.GetCameraImageURL(c, camera.ID)

	return c.JSON(fiber.Map{
		"error": false,
		"data":  camera,
	})
}

// CreateCamera handles creating a new camera
func (h *CameraHandler) CreateCamera(c *fiber.Ctx) error {
	ctx := c.Context()

	camera := new(entity.Camera)

	// Parse request body
	if err := c.BodyParser(camera); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Create camera
	err := h.cameraService.CreateCamera(ctx, camera)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for validation errors
		if err.Error() == "camera name is required" ||
			err.Error() == "camera location is required" ||
			err.Error() == "area ID is required" ||
			err.Error() == "invalid status. Must be active, inactive, maintenance, or issue" {
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
		"msg":   "Camera created successfully",
		"data":  camera,
	})
}

// UpdateCamera handles updating a camera
func (h *CameraHandler) UpdateCamera(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse ID parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid camera ID",
		})
	}

	// Parse request body
	updatedCamera := new(entity.Camera)
	if err := c.BodyParser(updatedCamera); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Set ID from URL parameter
	updatedCamera.ID = uint(id)

	// Update camera
	err = h.cameraService.UpdateCamera(ctx, updatedCamera)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for specific errors
		if err.Error() == "camera not found" || err.Error() == "area not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "invalid status. Must be active, inactive, maintenance, or issue" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get updated camera
	camera, err := h.cameraService.GetCameraByID(ctx, uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Error retrieving updated camera: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   "Camera updated successfully",
		"data":  camera,
	})
}

// DeleteCamera handles deleting a camera
func (h *CameraHandler) DeleteCamera(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse ID parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid camera ID",
		})
	}

	// Delete camera
	err = h.cameraService.DeleteCamera(ctx, uint(id))
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "camera not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   "Camera deleted successfully",
	})
}

// UpdateCameraStatus handles updating just the status of a camera
func (h *CameraHandler) UpdateCameraStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse ID parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid camera ID",
		})
	}

	// Parse request body
	type StatusUpdate struct {
		Status string `json:"status"`
	}

	statusUpdate := new(StatusUpdate)
	if err := c.BodyParser(statusUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Update camera status
	err = h.cameraService.UpdateCameraStatus(ctx, uint(id), statusUpdate.Status)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "camera not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "invalid status. Must be active, inactive, maintenance, or issue" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get updated camera
	camera, err := h.cameraService.GetCameraByID(ctx, uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Error retrieving updated camera: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   "Camera status updated successfully",
		"data":  camera,
	})
}
