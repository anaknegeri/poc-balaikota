package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"
	"people-counting/internal/domain/service"
)

// parseFlexibleDate attempts to parse date in multiple formats
func parseFlexibleDate(dateStr string) (time.Time, error) {
	// Try different date formats
	formats := []string{
		time.RFC3339,          // 2025-07-23T10:00:00Z
		"2006-01-02",          // 2025-07-23
		"2006-01-02T15:04:05", // 2025-07-23T10:00:00
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("invalid date format. Supported formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, or RFC3339")
}

// PeopleCountServiceImpl implements service.PeopleCountService
type PeopleCountServiceImpl struct {
	peopleCountRepository repository.PeopleCountRepository
}

// NewPeopleCountService creates a new people count service
func NewPeopleCountService(
	peopleCountRepository repository.PeopleCountRepository,
) service.PeopleCountService {
	return &PeopleCountServiceImpl{
		peopleCountRepository: peopleCountRepository,
	}
}

func (s *PeopleCountServiceImpl) GetAlertByID(ctx context.Context, id string) (*entity.PeopleCount, error) {
	if id == "" {
		return nil, errors.New("alert ID is required")
	}

	return s.peopleCountRepository.FindByID(ctx, id)
}

// GetAllCounts retrieves paginated people count records
func (s *PeopleCountServiceImpl) GetAllCounts(ctx context.Context, page, limit int, areaID, from, to string, includeArea bool) ([]entity.PeopleCount, int64, error) {
	// Use default pagination values if invalid
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 50
	}

	// Prepare filters
	filters := make(map[string]interface{})

	// Add area filter if provided
	if areaID != "" {
		id, err := strconv.ParseUint(areaID, 10, 64)
		if err != nil {
			return nil, 0, errors.New("invalid area ID")
		}

		filters["area_id"] = uint(id)
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

	// Add include area flag if requested
	if includeArea {
		filters["include_area"] = true
	}

	return s.peopleCountRepository.FindAll(ctx, page, limit, filters)
}

func (s *PeopleCountServiceImpl) CreatePeopleCount(ctx context.Context, counting *entity.PeopleCount) error {
	return s.peopleCountRepository.Create(ctx, counting)
}

func (s *PeopleCountServiceImpl) UpdatePeopleCount(ctx context.Context, counting *entity.PeopleCount) error {
	// Validate required fields
	if counting.ID == "" {
		return errors.New("people count ID is required")
	}

	return s.peopleCountRepository.Update(ctx, counting)
}

func (s *PeopleCountServiceImpl) GetByID(ctx context.Context, id string) (*entity.PeopleCount, error) {
	if id == "" {
		return nil, errors.New("people count ID is required")
	}

	return s.peopleCountRepository.FindByID(ctx, id)
}

// RecordCount creates a new people count record
func (s *PeopleCountServiceImpl) RecordCount(ctx context.Context, count *entity.PeopleCount) error {
	// Ensure counts are not negative
	if count.MaleCount < 0 {
		count.MaleCount = 0
	}

	if count.FemaleCount < 0 {
		count.FemaleCount = 0
	}

	if count.ChildCount < 0 {
		count.ChildCount = 0
	}

	if count.AdultCount < 0 {
		count.AdultCount = 0
	}

	if count.ElderlyCount < 0 {
		count.ElderlyCount = 0
	}

	// Check demographic consistency
	demographicSum := count.ChildCount + count.AdultCount + count.ElderlyCount
	genderSum := count.MaleCount + count.FemaleCount

	if demographicSum != genderSum && demographicSum != 0 && genderSum != 0 {
		return errors.New("demographic counts (child + adult + elderly) must equal gender counts (male + female)")
	}

	// Set timestamp to current time if not provided
	if count.Timestamp.IsZero() {
		count.Timestamp = time.Now()
	}

	return s.peopleCountRepository.Create(ctx, count)
}

// GetCountsSummary retrieves a summary of current people counts
func (s *PeopleCountServiceImpl) GetCountsSummary(ctx context.Context, from, to string) (*entity.CountSummary, error) {
	filters := make(map[string]interface{})
	if from != "" {
		fromTime, err := parseFlexibleDate(from)
		if err != nil {
			return nil, errors.New("invalid 'from' date format. " + err.Error())
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := parseFlexibleDate(to)
		if err != nil {
			return nil, errors.New("invalid 'to' date format. " + err.Error())
		}
		filters["to"] = toTime
	}

	return s.peopleCountRepository.GetSummary(ctx, filters)
}

// GetCountsTrend retrieves trend data for people counts
func (s *PeopleCountServiceImpl) GetCountsTrend(ctx context.Context, interval string, areaID, from, to string) (*entity.CountsByTimeResult, error) {
	// Validate interval
	validIntervals := map[string]bool{
		"hour":  true,
		"day":   true,
		"week":  true,
		"month": true,
	}

	if !validIntervals[interval] {
		return nil, errors.New("invalid interval. Must be hour, day, week, or month")
	}

	// Process area filter if provided

	filters := make(map[string]interface{})
	if areaID != "" {
		id, err := strconv.ParseUint(areaID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid area ID")
		}

		areaIDUint := uint(id)
		filters["area_id"] = &areaIDUint
	}

	if from != "" {
		fromTime, err := parseFlexibleDate(from)
		if err != nil {
			return nil, errors.New("invalid 'from' date format. " + err.Error())
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := parseFlexibleDate(to)
		if err != nil {
			return nil, errors.New("invalid 'to' date format. " + err.Error())
		}
		filters["to"] = toTime
	}

	return s.peopleCountRepository.GetTrends(ctx, interval, filters)
}

// GetCountsDistribution retrieves distribution data
func (s *PeopleCountServiceImpl) GetCountsDistribution(ctx context.Context, distType, timeWindow string) (interface{}, error) {
	// Validate distribution type
	if distType != "camera" && distType != "gender" && distType != "age" {
		return nil, errors.New("invalid distribution type. Must be camera, gender, or age")
	}

	// Parse time window
	var timeConstraint time.Time

	switch timeWindow {
	case "24h":
		timeConstraint = time.Now().Add(-24 * time.Hour)
	case "7d":
		timeConstraint = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		timeConstraint = time.Now().Add(-30 * 24 * time.Hour)
	case "all":
		timeConstraint = time.Time{} // Zero time to get all data
	default:
		timeConstraint = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}

	// Get distribution data based on type
	switch distType {
	case "camera":
		cameras, err := s.peopleCountRepository.GetDistributionByCamera(ctx, timeConstraint)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"type": "camera",
			"data": cameras,
		}, nil

	case "gender":
		counts, err := s.peopleCountRepository.GetDistributionByGender(ctx, timeConstraint)
		if err != nil {
			return nil, err
		}

		// Calculate percentages
		malePercentage := 0
		femalePercentage := 0

		if counts.Total > 0 {
			malePercentage = int((float64(counts.Male) / float64(counts.Total)) * 100)
			femalePercentage = 100 - malePercentage // Ensure they add up to 100
		}

		return map[string]interface{}{
			"type": "gender",
			"data": map[string]interface{}{
				"counts": counts,
				"percentages": map[string]int{
					"male":   malePercentage,
					"female": femalePercentage,
				},
			},
		}, nil

	case "age":
		counts, err := s.peopleCountRepository.GetDistributionByAge(ctx, timeConstraint)
		if err != nil {
			return nil, err
		}

		// Calculate percentages
		childPercentage := 0
		adultPercentage := 0
		elderlyPercentage := 0

		if counts.Total > 0 {
			childPercentage = int((float64(counts.Child) / float64(counts.Total)) * 100)
			adultPercentage = int((float64(counts.Adult) / float64(counts.Total)) * 100)
			elderlyPercentage = int((float64(counts.Elderly) / float64(counts.Total)) * 100)

			// Adjust to ensure they sum to 100
			sum := childPercentage + adultPercentage + elderlyPercentage
			if sum != 100 {
				// Add the remainder to the largest group
				if counts.Child >= counts.Adult && counts.Child >= counts.Elderly {
					childPercentage += (100 - sum)
				} else if counts.Adult >= counts.Child && counts.Adult >= counts.Elderly {
					adultPercentage += (100 - sum)
				} else {
					elderlyPercentage += (100 - sum)
				}
			}
		}

		return map[string]interface{}{
			"type": "age",
			"data": map[string]interface{}{
				"counts": counts,
				"percentages": map[string]int{
					"child":   childPercentage,
					"adult":   adultPercentage,
					"elderly": elderlyPercentage,
				},
			},
		}, nil
	}

	// Should never reach here due to validation above
	return nil, errors.New("invalid distribution type")
}

// GetPeakHoursAnalysis retrieves peak hours analysis
func (s *PeopleCountServiceImpl) GetPeakHoursAnalysis(ctx context.Context, cameraID string, from, to string) (*entity.PeakHoursAnalysis, error) {
	// Process camera filter if provided
	filters := make(map[string]interface{})
	if cameraID != "" {
		id, err := strconv.ParseUint(cameraID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid area ID")
		}

		cameraIDUint := uint(id)
		filters["camera_id"] = &cameraIDUint
	}

	if from != "" {
		fromTime, err := parseFlexibleDate(from)
		if err != nil {
			return nil, errors.New("invalid 'from' date format. " + err.Error())
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := parseFlexibleDate(to)
		if err != nil {
			return nil, errors.New("invalid 'to' date format. " + err.Error())
		}
		filters["to"] = toTime
	}

	return s.peopleCountRepository.GetPeakHoursAnalysis(ctx, filters)
}
