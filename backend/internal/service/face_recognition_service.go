package service

import (
	"context"
	"errors"
	"strconv"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"
	"people-counting/internal/domain/service"
)

// FaceRecognitionServiceImpl implements service.FaceRecognition
type FaceRecognitionServiceImpl struct {
	faceRecognitionRepository repository.FaceRecognitionRepository
	cameraRepository          repository.CameraRepository
}

// NewFaceRecognitionService creates a new face recognition service
func NewFaceRecognitionService(
	faceRecognitionRepository repository.FaceRecognitionRepository,
	cameraRepository repository.CameraRepository,
) service.FaceRecognitionService {
	return &FaceRecognitionServiceImpl{
		faceRecognitionRepository: faceRecognitionRepository,
		cameraRepository:          cameraRepository,
	}
}

// GetAll retrieves paginated face recognition records with filters
func (s *FaceRecognitionServiceImpl) GetAll(ctx context.Context, page, limit int, cameraID, from, to string, includeRelations bool) ([]entity.FaceRecognition, int64, error) {
	// Use default pagination values if invalid
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 50
	}

	// Prepare filters
	filters := make(map[string]interface{})

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

	return s.faceRecognitionRepository.FindAll(ctx, page, limit, filters)
}

// Create creates a new face recognition record
func (s *FaceRecognitionServiceImpl) Create(ctx context.Context, data *entity.FaceRecognition) error {
	// Verify camera exists if provided
	_, err := s.cameraRepository.FindByID(ctx, data.CameraID)
	if err != nil {
		return errors.New("camera not found")
	}

	// Create face recognition record
	return s.faceRecognitionRepository.Create(ctx, data)
}

// Update updates an existing face recognition record
func (s *FaceRecognitionServiceImpl) Update(ctx context.Context, data *entity.FaceRecognition) error {
	// Validate data
	if data.ID == "" {
		return errors.New("ID is required")
	}

	// Verify camera exists if provided
	_, err := s.cameraRepository.FindByID(ctx, data.CameraID)
	if err != nil {
		return errors.New("camera not found")
	}

	// Update face recognition record
	return s.faceRecognitionRepository.Update(ctx, data)
}

// GetByID retrieves a face recognition record by ID
func (s *FaceRecognitionServiceImpl) GetByID(ctx context.Context, id string) (*entity.FaceRecognition, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.faceRecognitionRepository.FindByID(ctx, id)
}
