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

// VehicleCountServiceImpl implements service.VehicleCountService
type VehicleCountServiceImpl struct {
	vehicleCountRepository repository.VehicleCountRepository
}

// NewVehicleCountService creates a new vehicle count service
func NewVehicleCountService(
	vehicleCountRepository repository.VehicleCountRepository,
) service.VehicleCountService {
	return &VehicleCountServiceImpl{
		vehicleCountRepository: vehicleCountRepository,
	}
}

func (s *VehicleCountServiceImpl) GetVehicleCountByID(ctx context.Context, id string) (*entity.VehicleCount, error) {
	if id == "" {
		return nil, errors.New("vehicle count ID is required")
	}

	return s.vehicleCountRepository.FindByID(ctx, id)
}

// GetAllCounts retrieves paginated vehicle count records
func (s *VehicleCountServiceImpl) GetAllCounts(ctx context.Context, page, limit int, cctvID, from, to string, includeCctv bool) ([]entity.VehicleCount, int64, error) {
	// Use default pagination values if invalid
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 50
	}

	// Prepare filters
	filters := make(map[string]interface{})

	// Add CCTV filter if provided
	if cctvID != "" {
		id, err := strconv.ParseUint(cctvID, 10, 64)
		if err != nil {
			return nil, 0, errors.New("invalid cctv ID")
		}

		filters["cctv_id"] = uint(id)
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

	// Add include CCTV flag if requested
	if includeCctv {
		filters["include_cctv"] = true
	}

	return s.vehicleCountRepository.FindAll(ctx, page, limit, filters)
}

func (s *VehicleCountServiceImpl) CreateVehicleCount(ctx context.Context, count *entity.VehicleCount) error {
	return s.vehicleCountRepository.Create(ctx, count)
}

func (s *VehicleCountServiceImpl) UpdateVehicleCount(ctx context.Context, count *entity.VehicleCount) error {
	// Validate required fields
	if count.ID == "" {
		return errors.New("vehicle count ID is required")
	}

	// Validate CCTV ID
	if count.CctvID == 0 {
		return errors.New("cctv ID is required")
	}

	return s.vehicleCountRepository.Update(ctx, count)
}

// RecordCount creates a new vehicle count record
func (s *VehicleCountServiceImpl) RecordCount(ctx context.Context, count *entity.VehicleCount) error {
	// Validate CCTV ID
	if count.CctvID == 0 {
		return errors.New("cctv ID is required")
	}

	// Ensure counts are not negative
	if count.InCountCar < 0 {
		count.InCountCar = 0
	}

	if count.InCountTruck < 0 {
		count.InCountTruck = 0
	}

	if count.InCountPeople < 0 {
		count.InCountPeople = 0
	}

	if count.OutCount < 0 {
		count.OutCount = 0
	}

	// Set timestamp to current time if not provided
	if count.Timestamp.IsZero() {
		count.Timestamp = time.Now()
	}

	// Set device timestamp if not provided
	if count.DeviceTimestamp.IsZero() {
		count.DeviceTimestamp = count.Timestamp
	}

	// Set device timestamp UTC if not provided
	if count.DeviceTimestampUTC == 0 {
		count.DeviceTimestampUTC = float64(count.DeviceTimestamp.Unix())
	}

	return s.vehicleCountRepository.Create(ctx, count)
}

// GetCountsSummary retrieves a summary of current vehicle counts
func (s *VehicleCountServiceImpl) GetCountsSummary(ctx context.Context, cctvID, from, to string) (*entity.VehicleCountSummary, error) {
	filters := make(map[string]interface{})
	if cctvID != "" {
		id, err := strconv.ParseUint(cctvID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid area ID")
		}

		cctvIDUint := uint(id)
		filters["cctv_id"] = &cctvIDUint
	}

	if from != "" {
		fromTime, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return nil, errors.New("invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)")
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return nil, errors.New("invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)")
		}
		filters["to"] = toTime
	}

	return s.vehicleCountRepository.GetSummary(ctx, filters)
}

// GetCountsTrend retrieves trend data for vehicle counts
func (s *VehicleCountServiceImpl) GetCountsTrend(ctx context.Context, interval string, cctvID, from, to string) (*entity.VehicleCountsByTimeResult, error) {
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

	// Process CCTV filter if provided
	filters := make(map[string]interface{})
	if cctvID != "" {
		id, err := strconv.ParseUint(cctvID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid area ID")
		}

		cctvIDUint := uint(id)
		filters["cctv_id"] = &cctvIDUint
	}

	if from != "" {
		fromTime, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return nil, errors.New("invalid 'from' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)")
		}
		filters["from"] = fromTime
	}

	if to != "" {
		toTime, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return nil, errors.New("invalid 'to' date format. Use RFC3339 format (e.g. 2025-05-13T10:00:00Z)")
		}
		filters["to"] = toTime
	}

	return s.vehicleCountRepository.GetTrends(ctx, interval, filters)
}

// GetCountsDistribution retrieves distribution data
func (s *VehicleCountServiceImpl) GetCountsDistribution(ctx context.Context, distType, timeWindow string) (interface{}, error) {
	// Validate distribution type
	if distType != "cctv" && distType != "vehicle_type" {
		return nil, errors.New("invalid distribution type. Must be cctv or vehicle_type")
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
	case "cctv":
		cctvs, err := s.vehicleCountRepository.GetDistributionByCctv(ctx, timeConstraint)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"type": "cctv",
			"data": cctvs,
		}, nil

	case "vehicle_type":
		counts, err := s.vehicleCountRepository.GetDistributionByVehicleType(ctx, timeConstraint)
		if err != nil {
			return nil, err
		}

		// Calculate percentages for vehicle types
		carPercentage := 0
		truckPercentage := 0
		peoplePercentage := 0

		if counts.TotalVehicleIn > 0 {
			carPercentage = int((float64(counts.InCar) / float64(counts.TotalVehicleIn)) * 100)
			truckPercentage = int((float64(counts.InTruck) / float64(counts.TotalVehicleIn)) * 100)

			// Adjust to ensure they sum to 100 for vehicles
			remainder := 100 - carPercentage - truckPercentage
			if remainder != 0 {
				if counts.InCar >= counts.InTruck {
					carPercentage += remainder
				} else {
					truckPercentage += remainder
				}
			}
		}

		// Calculate people percentage against total in count
		if counts.TotalIn > 0 {
			peoplePercentage = int((float64(counts.InPeople) / float64(counts.TotalIn)) * 100)
		}

		return map[string]interface{}{
			"type": "vehicle_type",
			"data": map[string]interface{}{
				"counts": counts,
				"percentages": map[string]interface{}{
					"vehicles": map[string]int{
						"car":   carPercentage,
						"truck": truckPercentage,
					},
					"people": peoplePercentage,
				},
			},
		}, nil
	}

	// Should never reach here due to validation above
	return nil, errors.New("invalid distribution type")
}

// GetPeakHours retrieves peak hours analysis for vehicle counts
func (s *VehicleCountServiceImpl) GetPeakHours(ctx context.Context, cctvID string, days int) ([]entity.VehicleTrendPoint, error) {
	// Use default days if invalid
	if days <= 0 {
		days = 7 // Default to last 7 days
	}

	// Process CCTV filter if provided
	var cctvIDPtr *uint
	if cctvID != "" {
		id, err := strconv.ParseUint(cctvID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid cctv ID")
		}

		cctvIDUint := uint(id)
		cctvIDPtr = &cctvIDUint
	}

	return s.vehicleCountRepository.GetPeakHours(ctx, cctvIDPtr, days)
}

// GetLatestByCctv retrieves the latest vehicle count for a specific CCTV
func (s *VehicleCountServiceImpl) GetLatestByCctv(ctx context.Context, cctvIDStr string) (*entity.VehicleCount, error) {
	if cctvIDStr == "" {
		return nil, errors.New("cctv ID is required")
	}

	cctvID, err := strconv.ParseUint(cctvIDStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid cctv ID")
	}

	return s.vehicleCountRepository.GetLatestByCctv(ctx, uint(cctvID))
}

// GetCountsByTimeRange retrieves vehicle counts within a specific time range
func (s *VehicleCountServiceImpl) GetCountsByTimeRange(ctx context.Context, from, to time.Time, cctvID string) ([]entity.VehicleCount, error) {
	// Process CCTV filter if provided
	var cctvIDPtr *uint
	if cctvID != "" {
		id, err := strconv.ParseUint(cctvID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid cctv ID")
		}

		cctvIDUint := uint(id)
		cctvIDPtr = &cctvIDUint
	}

	return s.vehicleCountRepository.GetCountsByTimeRange(ctx, from, to, cctvIDPtr)
}

// CalculateOccupancyRate calculates the occupancy rate for a given time period
func (s *VehicleCountServiceImpl) CalculateOccupancyRate(ctx context.Context, cctvID string, from, to time.Time) (float64, error) {
	counts, err := s.GetCountsByTimeRange(ctx, from, to, cctvID)
	if err != nil {
		return 0, err
	}

	if len(counts) == 0 {
		return 0, nil
	}

	totalIn := 0
	totalOut := 0

	for _, count := range counts {
		totalIn += count.TotalInCount
		totalOut += count.OutCount
	}

	// Simple occupancy calculation: (total_in - total_out) / time_period_hours
	duration := to.Sub(from).Hours()
	if duration == 0 {
		return 0, nil
	}

	netCount := totalIn - totalOut
	occupancyRate := float64(netCount) / duration

	// Ensure non-negative rate
	if occupancyRate < 0 {
		occupancyRate = 0
	}

	return occupancyRate, nil
}

// GetBusiestCctvs retrieves the busiest CCTVs based on vehicle count
func (s *VehicleCountServiceImpl) GetBusiestCctvs(ctx context.Context, timeWindow time.Time, limit int) ([]entity.CctvVehicleSummary, error) {
	if limit <= 0 {
		limit = 10 // Default to top 10
	}

	distribution, err := s.vehicleCountRepository.GetDistributionByCctv(ctx, timeWindow)
	if err != nil {
		return nil, err
	}

	// Limit results
	if len(distribution) > limit {
		distribution = distribution[:limit]
	}

	return distribution, nil
}
