package postgres

import (
	"context"
	"errors"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"

	"gorm.io/gorm"
)

// VehicleCountRepositoryImpl implements repository.VehicleCountRepository
type VehicleCountRepositoryImpl struct {
	db *gorm.DB
}

// NewVehicleCountRepository creates a new vehicle count repository
func NewVehicleCountRepository(db *gorm.DB) repository.VehicleCountRepository {
	return &VehicleCountRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves paginated vehicle count records with filters
func (r *VehicleCountRepositoryImpl) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.VehicleCount, int64, error) {
	var counts []entity.VehicleCount
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Build query
	query := r.db.WithContext(ctx).Model(&entity.VehicleCount{}).Order("timestamp DESC")

	// Apply filters if provided
	if filters != nil {
		if cctvID, ok := filters["cctv_id"].(uint); ok && cctvID != 0 {
			query = query.Where("cctv_id = ?", cctvID)
		}

		if from, ok := filters["from"].(time.Time); ok && !from.IsZero() {
			query = query.Where("timestamp >= ?", from)
		}

		if to, ok := filters["to"].(time.Time); ok && !to.IsZero() {
			query = query.Where("timestamp <= ?", to)
		}

		if includeCctv, ok := filters["include_cctv"].(bool); ok && includeCctv {
			query = query.Preload("Cctv")
		}
	}

	// Get total count for pagination
	countQuery := query
	countQuery.Count(&total)

	// Get paginated results
	result := query.Limit(limit).Offset(offset).Find(&counts)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Calculate total counts for each record
	for i := range counts {
		counts[i].CalculateTotalCounts()
	}

	return counts, total, nil
}

// FindByID finds a vehicle count record by its ID
func (r *VehicleCountRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.VehicleCount, error) {
	var count entity.VehicleCount

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&count)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("vehicle count record not found")
		}
		return nil, result.Error
	}

	// Calculate total counts
	count.CalculateTotalCounts()

	return &count, nil
}

// FindByCctv finds vehicle count records for a specific CCTV
func (r *VehicleCountRepositoryImpl) FindByCctv(ctx context.Context, cctvID uint, from, to time.Time, limit int) ([]entity.VehicleCount, error) {
	var counts []entity.VehicleCount

	query := r.db.WithContext(ctx).
		Where("cctv_id = ?", cctvID).
		Order("timestamp DESC").
		Limit(limit)

	// Apply time range if provided
	if !from.IsZero() && !to.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", from, to)
	} else if !from.IsZero() {
		query = query.Where("timestamp >= ?", from)
	} else if !to.IsZero() {
		query = query.Where("timestamp <= ?", to)
	}

	result := query.Find(&counts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Calculate total counts for each record
	for i := range counts {
		counts[i].CalculateTotalCounts()
	}

	return counts, nil
}

// Create adds a new vehicle count record to the database
func (r *VehicleCountRepositoryImpl) Create(ctx context.Context, count *entity.VehicleCount) error {
	// Set timestamp to current time if not provided
	if count.Timestamp.IsZero() {
		count.Timestamp = time.Now()
	}

	// Calculate total counts before saving
	count.CalculateTotalCounts()

	result := r.db.WithContext(ctx).Create(count)
	return result.Error
}

// Update updates an existing vehicle count record in the database
func (r *VehicleCountRepositoryImpl) Update(ctx context.Context, count *entity.VehicleCount) error {
	// Check if the record exists
	var existingCount entity.VehicleCount
	result := r.db.WithContext(ctx).Where("id = ?", count.ID).First(&existingCount)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("vehicle count record not found")
		}
		return result.Error
	}

	// Calculate total counts before updating
	count.CalculateTotalCounts()

	// Update the record
	result = r.db.WithContext(ctx).Model(&existingCount).Updates(count)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetSummary retrieves a summary of current vehicle counts
func (r *VehicleCountRepositoryImpl) GetSummary(ctx context.Context, filters map[string]interface{}) (*entity.VehicleCountSummary, error) {
	// Initialize summary
	summary := &entity.VehicleCountSummary{
		Cctvs:  []entity.CctvVehicleSummary{},
		Totals: entity.VehicleTotalCounts{},
	}

	type Raw struct {
		CctvID              uint      `gorm:"column:cctv_id"`
		CctvName            string    `gorm:"column:cctv_name"`
		InCountCar          int       `gorm:"column:in_count_car"`
		InCountTruck        int       `gorm:"column:in_count_truck"`
		InCountPeople       int       `gorm:"column:in_count_people"`
		OutCount            int       `gorm:"column:out_count"`
		TotalInCount        int       `gorm:"column:total_in_count"`
		TotalVehicleInCount int       `gorm:"column:total_vehicle_in_count"`
		Timestamp           time.Time `gorm:"column:timestamp"`
	}

	var rows []Raw

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	if filters != nil {
		if from, ok := filters["from"].(time.Time); ok && !from.IsZero() {
			startOfDay = from
		}
		if to, ok := filters["to"].(time.Time); ok && !to.IsZero() {
			endOfDay = to
		}
	}

	// Get all vehicle_counts data joined with cameras
	err := r.db.WithContext(ctx).
		Table("vehicle_counts vc").
		Select("vc.cctv_id, c.name as cctv_name, vc.in_count_car, vc.in_count_truck, vc.in_count_people, vc.out_count, vc.total_in_count, vc.total_vehicle_in_count, vc.timestamp").
		Joins("JOIN cameras c ON vc.cctv_id = c.id").
		Where("vc.timestamp >= ? AND  vc.timestamp <= ?", startOfDay, endOfDay).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	// Group by cctv_id
	cctvMap := map[uint]*entity.CctvVehicleSummary{}

	for _, row := range rows {
		cctv, exists := cctvMap[row.CctvID]
		if !exists {
			cctv = &entity.CctvVehicleSummary{
				CctvID:      row.CctvID,
				CctvName:    row.CctvName,
				LastUpdated: row.Timestamp,
			}
			cctvMap[row.CctvID] = cctv
		}

		// Accumulate per CCTV
		cctv.InCountCar += row.InCountCar
		cctv.InCountTruck += row.InCountTruck
		cctv.InCountPeople += row.InCountPeople
		cctv.OutCount += row.OutCount
		cctv.TotalInCount += row.TotalInCount
		cctv.TotalVehicleInCount += row.TotalVehicleInCount
		cctv.NetCount = cctv.TotalInCount - cctv.OutCount

		// Update last updated time
		if row.Timestamp.After(cctv.LastUpdated) {
			cctv.LastUpdated = row.Timestamp
		}

		// Accumulate overall totals
		summary.Totals.InCar += row.InCountCar
		summary.Totals.InTruck += row.InCountTruck
		summary.Totals.InPeople += row.InCountPeople
		summary.Totals.Out += row.OutCount
		summary.Totals.TotalIn += row.TotalInCount
		summary.Totals.TotalVehicleIn += row.TotalVehicleInCount
	}

	// Calculate overall net count
	summary.Totals.NetCount = summary.Totals.TotalIn - summary.Totals.Out

	// Convert map to slice
	for _, cctv := range cctvMap {
		summary.Cctvs = append(summary.Cctvs, *cctv)
	}

	return summary, nil
}

// GetTrends retrieves trend data for vehicle counts
func (r *VehicleCountRepositoryImpl) GetTrends(ctx context.Context, interval string, filters map[string]interface{}) (*entity.VehicleCountsByTimeResult, error) {
	// SQL with appropriate time truncation based on interval
	var timeFormat string
	switch interval {
	case "hour":
		timeFormat = "date_trunc('hour', timestamp)"
	case "day":
		timeFormat = "date_trunc('day', timestamp)"
	case "week":
		timeFormat = "date_trunc('week', timestamp)"
	case "month":
		timeFormat = "date_trunc('month', timestamp)"
	default:
		timeFormat = "date_trunc('hour', timestamp)"
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	if filters != nil {
		if from, ok := filters["from"].(time.Time); ok && !from.IsZero() {
			startOfDay = from
		}
		if to, ok := filters["to"].(time.Time); ok && !to.IsZero() {
			endOfDay = to
		}
	}

	// Build base query
	query := r.db.WithContext(ctx).Table("vehicle_counts").
		Select(timeFormat+" as time_period, SUM(in_count_car) as in_count_car, SUM(in_count_truck) as in_count_truck, SUM(in_count_people) as in_count_people, SUM(out_count) as out_count, SUM(total_in_count) as total_in_count, SUM(total_vehicle_in_count) as total_vehicle_in_count, (SUM(total_in_count) - SUM(out_count)) as net_count").
		Group("time_period").
		Where("timestamp >= ? AND timestamp <= ?", startOfDay, endOfDay).
		Order("time_period DESC")

	if filters != nil {
		if cctvID, ok := filters["cctv_id"].(uint); ok && cctvID > 0 {
			query = query.Where("cctv_id = ?", cctvID)
		}
		if cctvIDStr, ok := filters["cctv_id"].(string); ok && cctvIDStr != "" {
			query = query.Where("cctv_id = ?", cctvIDStr)
		}
	}

	// Execute query
	var trends []entity.VehicleTrendPoint

	err := query.Find(&trends).Error
	if err != nil {
		return nil, err
	}

	return &entity.VehicleCountsByTimeResult{
		Interval: interval,
		Data:     trends,
	}, nil
}

// GetDistributionByCctv retrieves distribution data by CCTV
func (r *VehicleCountRepositoryImpl) GetDistributionByCctv(ctx context.Context, timeWindow time.Time) ([]entity.CctvVehicleSummary, error) {
	var distribution []entity.CctvVehicleSummary

	query := r.db.WithContext(ctx).Table("vehicle_counts vc").
		Select("vc.cctv_id, c.name as cctv_name, SUM(vc.in_count_car) as in_count_car, SUM(vc.in_count_truck) as in_count_truck, SUM(vc.in_count_people) as in_count_people, SUM(vc.out_count) as out_count, SUM(vc.total_in_count) as total_in_count, SUM(vc.total_vehicle_in_count) as total_vehicle_in_count, (SUM(vc.total_in_count) - SUM(vc.out_count)) as net_count, MAX(vc.timestamp) as last_updated").
		Joins("JOIN cameras c ON vc.cctv_id = c.id")

	// Apply time constraint if not zero time
	if !timeWindow.IsZero() {
		query = query.Where("vc.timestamp >= ?", timeWindow)
	}

	err := query.Group("vc.cctv_id, c.name").
		Order("total_in_count DESC").
		Find(&distribution).Error

	if err != nil {
		return nil, err
	}

	return distribution, nil
}

// GetDistributionByVehicleType retrieves distribution data by vehicle type
func (r *VehicleCountRepositoryImpl) GetDistributionByVehicleType(ctx context.Context, timeWindow time.Time) (*entity.VehicleTotalCounts, error) {
	var counts entity.VehicleTotalCounts

	query := r.db.WithContext(ctx).Table("vehicle_counts")

	// Apply time constraint if not zero time
	if !timeWindow.IsZero() {
		query = query.Where("timestamp >= ?", timeWindow)
	}

	err := query.Select("SUM(in_count_car) as in_car, SUM(in_count_truck) as in_truck, SUM(in_count_people) as in_people, SUM(out_count) as out, SUM(total_in_count) as total_in, SUM(total_vehicle_in_count) as total_vehicle_in, (SUM(total_in_count) - SUM(out_count)) as net_count").
		Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	return &counts, nil
}

// GetLatestByCctv retrieves the latest vehicle count record for a specific CCTV
func (r *VehicleCountRepositoryImpl) GetLatestByCctv(ctx context.Context, cctvID uint) (*entity.VehicleCount, error) {
	var count entity.VehicleCount

	result := r.db.WithContext(ctx).
		Where("cctv_id = ?", cctvID).
		Order("timestamp DESC").
		First(&count)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("no vehicle count record found for this CCTV")
		}
		return nil, result.Error
	}

	// Calculate total counts
	count.CalculateTotalCounts()

	return &count, nil
}

// GetCountsByTimeRange retrieves vehicle counts within a specific time range
func (r *VehicleCountRepositoryImpl) GetCountsByTimeRange(ctx context.Context, from, to time.Time, cctvID *uint) ([]entity.VehicleCount, error) {
	var counts []entity.VehicleCount

	query := r.db.WithContext(ctx).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("timestamp ASC")

	// Apply CCTV filter if provided
	if cctvID != nil {
		query = query.Where("cctv_id = ?", *cctvID)
	}

	result := query.Find(&counts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Calculate total counts for each record
	for i := range counts {
		counts[i].CalculateTotalCounts()
	}

	return counts, nil
}

// GetPeakHours retrieves peak hours analysis for vehicle counts
func (r *VehicleCountRepositoryImpl) GetPeakHours(ctx context.Context, cctvID *uint, days int) ([]entity.VehicleTrendPoint, error) {
	var trends []entity.VehicleTrendPoint

	query := r.db.WithContext(ctx).Table("vehicle_counts").
		Select("date_trunc('hour', timestamp) as time_period, AVG(total_in_count) as total_in_count, AVG(total_vehicle_in_count) as total_vehicle_in_count, AVG(out_count) as out_count, AVG(in_count_car) as in_count_car, AVG(in_count_truck) as in_count_truck, AVG(in_count_people) as in_count_people, (AVG(total_in_count) - AVG(out_count)) as net_count").
		Where("timestamp >= ?", time.Now().AddDate(0, 0, -days)).
		Group("date_trunc('hour', timestamp)").
		Order("total_in_count DESC").
		Limit(24) // Top 24 hours

	// Apply CCTV filter if provided
	if cctvID != nil {
		query = query.Where("cctv_id = ?", *cctvID)
	}

	err := query.Find(&trends).Error
	if err != nil {
		return nil, err
	}

	return trends, nil
}
