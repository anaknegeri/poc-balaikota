package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"
	"people-counting/internal/domain/service"

	"github.com/gofiber/fiber/v2"
)

// CameraServiceImpl implements service.CameraService
type CameraServiceImpl struct {
	cameraRepository repository.CameraRepository
	streamService    *CameraStreamService
	dataDir          string
}

// NewCameraService creates a new camera service
func NewCameraService(
	cameraRepository repository.CameraRepository,
	dataDir string,
) service.CameraService {
	service := &CameraServiceImpl{
		cameraRepository: cameraRepository,
		dataDir:          dataDir,
	}

	// Create and initialize the streaming service
	streamService := NewCameraStreamService(service, dataDir)
	service.streamService = streamService

	// Start the streaming service
	if err := streamService.Start(); err != nil {
		fmt.Printf("WARNING: Failed to start camera streaming service: %v\n", err)
	}

	return service
}

// GetAllCameras retrieves all cameras with optional filters
func (s *CameraServiceImpl) GetAllCameras(ctx context.Context, status string) ([]entity.Camera, error) {
	filters := make(map[string]interface{})

	if status != "" {
		// Validate status value
		validStatuses := []string{"active", "inactive", "maintenance", "issue"}
		isValid := false
		for _, s := range validStatuses {
			if status == s {
				isValid = true
				break
			}
		}

		if !isValid {
			return nil, errors.New("invalid status. Must be active, inactive, maintenance, or issue")
		}

		filters["status"] = status
	}

	cameras, err := s.cameraRepository.FindAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return cameras, nil
}

// GetCameraByID retrieves a camera by its ID
func (s *CameraServiceImpl) GetCameraByID(ctx context.Context, id uint) (*entity.Camera, error) {
	if id == 0 {
		return nil, errors.New("invalid camera ID")
	}

	camera, err := s.cameraRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return camera, nil
}

func (s *CameraServiceImpl) GetCameraStreamURL(c *fiber.Ctx, cameraID uint) string {
	return s.streamService.GetCameraStreamURL(c, cameraID)
}

func (s *CameraServiceImpl) GetCameraImageURL(c *fiber.Ctx, cameraID uint) string {
	return s.streamService.GetCameraImageURL(c, cameraID)
}

// CreateCamera creates a new camera
func (s *CameraServiceImpl) CreateCamera(ctx context.Context, camera *entity.Camera) error {
	// Validate camera
	if camera.Name == "" {
		return errors.New("camera name is required")
	}

	// Set default status if not provided
	if camera.Status == "" {
		camera.Status = "active"
	} else {
		// Validate status
		validStatuses := []string{"active", "inactive", "maintenance", "issue"}
		isValid := false
		for _, s := range validStatuses {
			if camera.Status == s {
				isValid = true
				break
			}
		}

		if !isValid {
			return errors.New("invalid status. Must be active, inactive, maintenance, or issue")
		}
	}

	return s.cameraRepository.Create(ctx, camera)
}

// UpdateCamera updates an existing camera
func (s *CameraServiceImpl) UpdateCamera(ctx context.Context, camera *entity.Camera) error {
	if camera.ID == 0 {
		return errors.New("invalid camera ID")
	}

	// Fetch existing camera to update only provided fields
	existingCamera, err := s.cameraRepository.FindByID(ctx, camera.ID)
	if err != nil {
		return err
	}

	// Update fields if provided
	if camera.Name != "" {
		existingCamera.Name = camera.Name
	}

	if camera.Location != "" {
		existingCamera.Location = camera.Location
	}

	if camera.Status != "" {
		// Validate status
		validStatuses := []string{"active", "inactive", "maintenance", "issue"}
		isValid := false
		for _, s := range validStatuses {
			if camera.Status == s {
				isValid = true
				break
			}
		}

		if !isValid {
			return errors.New("invalid status. Must be active, inactive, maintenance, or issue")
		}

		existingCamera.Status = camera.Status
	}

	// Update last online time
	existingCamera.UpdatedAt = time.Now()

	return s.cameraRepository.Update(ctx, existingCamera)
}

// UpdateCameraStatus updates just the status of a camera
func (s *CameraServiceImpl) UpdateCameraStatus(ctx context.Context, id uint, status string) error {
	if id == 0 {
		return errors.New("invalid camera ID")
	}

	// Validate status
	validStatuses := []string{"active", "inactive", "maintenance", "issue"}
	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("invalid status. Must be active, inactive, maintenance, or issue")
	}

	return s.cameraRepository.UpdateStatus(ctx, id, status)
}

// DeleteCamera deletes a camera
func (s *CameraServiceImpl) DeleteCamera(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("invalid camera ID")
	}

	return s.cameraRepository.Delete(ctx, id)
}
