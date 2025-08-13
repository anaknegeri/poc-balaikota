package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/repository"

	"gorm.io/gorm"
)

type HourlyResult struct {
	Hour     int `gorm:"column:hour"`
	Visitors int `gorm:"column:visitors"`
}

type PeopleCountRepositoryImpl struct {
	db *gorm.DB
}

func NewPeopleCountRepository(db *gorm.DB) repository.PeopleCountRepository {
	return &PeopleCountRepositoryImpl{
		db: db,
	}
}

func (r *PeopleCountRepositoryImpl) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]entity.PeopleCount, int64, error) {
	var counts []entity.PeopleCount
	var total int64

	offset := (page - 1) * limit
	query := r.db.WithContext(ctx).Model(&entity.PeopleCount{}).Order("timestamp DESC")

	if filters != nil {
		if areaID, ok := filters["camera_id"].(uint); ok && areaID != 0 {
			query = query.Where("camera_id = ?", areaID)
		}

		if from, ok := filters["from"].(time.Time); ok && !from.IsZero() {
			query = query.Where("timestamp >= ?", from)
		}

		if to, ok := filters["to"].(time.Time); ok && !to.IsZero() {
			query = query.Where("timestamp <= ?", to)
		}

		if includeArea, ok := filters["include_area"].(bool); ok && includeArea {
			query = query.Preload("Area")
		}
	}

	countQuery := query
	countQuery.Count(&total)

	result := query.Limit(limit).Offset(offset).Find(&counts)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	for i := range counts {
		counts[i].CalculateTotalCount()
	}

	return counts, total, nil
}

func (r *PeopleCountRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.PeopleCount, error) {
	var count entity.PeopleCount

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&count)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("count record not found")
		}
		return nil, result.Error
	}

	count.CalculateTotalCount()
	return &count, nil
}

func (r *PeopleCountRepositoryImpl) FindByArea(ctx context.Context, areaID uint, from, to time.Time, limit int) ([]entity.PeopleCount, error) {
	var counts []entity.PeopleCount

	query := r.db.WithContext(ctx).
		Where("camera_id = ?", areaID).
		Order("timestamp DESC").
		Limit(limit)

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

	for i := range counts {
		counts[i].CalculateTotalCount()
	}

	return counts, nil
}

func (r *PeopleCountRepositoryImpl) Create(ctx context.Context, count *entity.PeopleCount) error {
	if count.Timestamp.IsZero() {
		count.Timestamp = time.Now()
	}

	count.CalculateTotalCount()
	result := r.db.WithContext(ctx).Create(count)
	return result.Error
}

func (r *PeopleCountRepositoryImpl) Update(ctx context.Context, count *entity.PeopleCount) error {
	// Check if the record exists
	var existingCount entity.PeopleCount
	result := r.db.WithContext(ctx).Where("id = ?", count.ID).First(&existingCount)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("people count record not found")
		}
		return result.Error
	}

	// Update the record
	count.CalculateTotalCount()
	result = r.db.WithContext(ctx).Model(&existingCount).Updates(count)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *PeopleCountRepositoryImpl) GetSummary(ctx context.Context, filters map[string]interface{}) (*entity.CountSummary, error) {
	summary := &entity.CountSummary{
		Cameras: []entity.CameraSummary{},
		Totals:  entity.TotalCounts{},
	}

	type Raw struct {
		CameraID     uint      `gorm:"column:camera_id"`
		CameraName   string    `gorm:"column:camera_name"`
		MaleCount    int       `gorm:"column:male_count"`
		FemaleCount  int       `gorm:"column:female_count"`
		ChildCount   int       `gorm:"column:child_count"`
		AdultCount   int       `gorm:"column:adult_count"`
		ElderlyCount int       `gorm:"column:elderly_count"`
		Timestamp    time.Time `gorm:"column:timestamp"`
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

	var rows []Raw
	query := r.db.WithContext(ctx).
		Table("people_counts pc").
		Select("pc.camera_id, a.name as camera_name, pc.male_count, pc.female_count, pc.child_count, pc.adult_count, pc.elderly_count, pc.timestamp").
		Joins("JOIN cameras a ON pc.camera_id = a.id").
		Where("pc.timestamp >= ? AND pc.timestamp <= ?", startOfDay, endOfDay)

	if filters != nil {
		if cameraID, ok := filters["camera_id"].(uint); ok && cameraID > 0 {
			query = query.Where("pc.camera_id = ?", cameraID)
		}
		if cameraIDStr, ok := filters["camera_id"].(string); ok && cameraIDStr != "" {
			query = query.Where("pc.camera_id = ?", cameraIDStr)
		}
	}

	err := query.Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch people count summary: %w", err)
	}

	cameraMap := map[uint]*entity.CameraSummary{}

	for _, row := range rows {
		cam, exists := cameraMap[row.CameraID]
		if !exists {
			cam = &entity.CameraSummary{
				CameraID:    row.CameraID,
				CameraName:  row.CameraName,
				LastUpdated: row.Timestamp,
			}
			cameraMap[row.CameraID] = cam
		}

		cam.MaleCount += row.MaleCount
		cam.FemaleCount += row.FemaleCount
		cam.ChildCount += row.ChildCount
		cam.AdultCount += row.AdultCount
		cam.ElderlyCount += row.ElderlyCount
		cam.TotalCount += row.MaleCount + row.FemaleCount

		if row.Timestamp.After(cam.LastUpdated) {
			cam.LastUpdated = row.Timestamp
		}

		summary.Totals.Male += row.MaleCount
		summary.Totals.Female += row.FemaleCount
		summary.Totals.Child += row.ChildCount
		summary.Totals.Adult += row.AdultCount
		summary.Totals.Elderly += row.ElderlyCount
		summary.Totals.Total += row.MaleCount + row.FemaleCount
	}

	for _, cam := range cameraMap {
		summary.Cameras = append(summary.Cameras, *cam)
	}

	return summary, nil
}

func (r *PeopleCountRepositoryImpl) GetTrends(ctx context.Context, interval string, filters map[string]interface{}) (*entity.CountsByTimeResult, error) {
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

	query := r.db.WithContext(ctx).Table("people_counts").
		Select(timeFormat+" as time_period, SUM(male_count) as male_count, SUM(female_count) as female_count, SUM(male_count + female_count) as total_count, SUM(child_count) as child_count, SUM(adult_count) as adult_count, SUM(elderly_count) as elderly_count").
		Where("timestamp >= ? AND timestamp <= ?", startOfDay, endOfDay).
		Group("time_period").
		Order("time_period ASC")

	if filters != nil {
		if cameraID, ok := filters["camera_id"].(uint); ok && cameraID > 0 {
			query = query.Where("camera_id = ?", cameraID)
		}
		if cameraIDStr, ok := filters["camera_id"].(string); ok && cameraIDStr != "" {
			query = query.Where("camera_id = ?", cameraIDStr)
		}
	}

	var trends []entity.TrendPoint
	err := query.Find(&trends).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch people count trends: %w", err)
	}

	return &entity.CountsByTimeResult{
		Interval: interval,
		Data:     trends,
	}, nil
}

func (r *PeopleCountRepositoryImpl) GetDistributionByCamera(ctx context.Context, timeWindow time.Time) ([]entity.CameraSummary, error) {
	var distribution []entity.CameraSummary

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	query := r.db.WithContext(ctx).Table("people_counts pc").
		Select("pc.camera_id, a.name as camera_name, SUM(pc.male_count) as male_count, SUM(pc.female_count) as female_count, SUM(pc.male_count + pc.female_count) as total_count, SUM(pc.child_count) as child_count, SUM(pc.adult_count) as adult_count, SUM(pc.elderly_count) as elderly_count, MAX(pc.timestamp) as last_updated").
		Where("pc.timestamp >= ? AND pc.timestamp <= ?", startOfDay, endOfDay).
		Joins("JOIN cameras a ON pc.camera_id = a.id")

	if !timeWindow.IsZero() {
		query = query.Where("pc.timestamp >= ?", timeWindow)
	}

	err := query.Group("pc.camera_id, a.name").
		Order("total_count DESC").
		Find(&distribution).Error

	if err != nil {
		return nil, err
	}

	return distribution, nil
}

func (r *PeopleCountRepositoryImpl) GetDistributionByGender(ctx context.Context, timeWindow time.Time) (*entity.TotalCounts, error) {
	var counts entity.TotalCounts

	query := r.db.WithContext(ctx).Table("people_counts")

	if !timeWindow.IsZero() {
		query = query.Where("timestamp >= ?", timeWindow)
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	err := query.Select("SUM(male_count) as male, SUM(female_count) as female, SUM(male_count + female_count) as total").
		Where("timestamp >= ? AND timestamp <= ?", startOfDay, endOfDay).
		Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	return &counts, nil
}

func (r *PeopleCountRepositoryImpl) GetDistributionByAge(ctx context.Context, timeWindow time.Time) (*entity.TotalCounts, error) {
	var counts entity.TotalCounts

	query := r.db.WithContext(ctx).Table("people_counts")

	if !timeWindow.IsZero() {
		query = query.Where("timestamp >= ?", timeWindow)
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	err := query.Select("SUM(child_count) as child, SUM(adult_count) as adult, SUM(elderly_count) as elderly, SUM(male_count + female_count) as total").
		Where("timestamp >= ? AND timestamp <= ?", startOfDay, endOfDay).
		Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	return &counts, nil
}

func (r *PeopleCountRepositoryImpl) GetPeakHoursAnalysis(ctx context.Context, filters map[string]interface{}) (*entity.PeakHoursAnalysis, error) {
	var results []HourlyResult

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

	query := r.db.WithContext(ctx).Table("people_counts").
		Select("EXTRACT(hour FROM timestamp) as hour, SUM(male_count + female_count) as visitors").
		Group("EXTRACT(hour FROM timestamp)").
		Where("timestamp >= ? AND timestamp <= ?", startOfDay, endOfDay).
		Order("hour")

	if filters != nil {
		if cameraID, ok := filters["camera_id"].(uint); ok && cameraID > 0 {
			query = query.Where("camera_id = ?", cameraID)
		}
		if cameraIDStr, ok := filters["camera_id"].(string); ok && cameraIDStr != "" {
			query = query.Where("camera_id = ?", cameraIDStr)
		}
	}

	err := query.Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch peak hours analysis: %w", err)
	}

	if len(results) == 0 {
		return &entity.PeakHoursAnalysis{
			Data:    []entity.PeakHourPoint{},
			Summary: entity.PeakHoursSummary{},
			Insights: entity.PeakHoursInsights{
				BusiestPeriod:            "No data available",
				QuietestPeriod:           "No data available",
				RecommendedStaffingHours: []string{},
				TrafficPattern:           "No data",
			},
		}, nil
	}

	totalVisitors := 0
	maxVisitors := 0
	minVisitors := results[0].Visitors
	peakHour := results[0].Hour
	lowHour := results[0].Hour

	for _, result := range results {
		totalVisitors += result.Visitors

		if result.Visitors > maxVisitors {
			maxVisitors = result.Visitors
			peakHour = result.Hour
		}

		if result.Visitors < minVisitors {
			minVisitors = result.Visitors
			lowHour = result.Hour
		}
	}

	averagePerHour := float64(totalVisitors) / float64(len(results))

	peakVsLowRatio := float64(0)
	if minVisitors > 0 {
		peakVsLowRatio = float64(maxVisitors) / float64(minVisitors)
	}

	var dataPoints []entity.PeakHourPoint
	for _, result := range results {
		percentage := 0
		if maxVisitors > 0 {
			percentage = int((float64(result.Visitors) / float64(maxVisitors)) * 100)
		}

		period := "low"
		if percentage >= 80 {
			period = "peak"
		} else if percentage >= 60 {
			period = "high"
		} else if percentage >= 40 {
			period = "moderate"
		}

		timeLabel := fmt.Sprintf("%02d:00", result.Hour)

		dataPoints = append(dataPoints, entity.PeakHourPoint{
			Hour:       result.Hour,
			TimeLabel:  timeLabel,
			Visitors:   result.Visitors,
			Period:     period,
			Percentage: percentage,
		})
	}

	summary := entity.PeakHoursSummary{
		PeakHour:       peakHour,
		PeakCount:      maxVisitors,
		LowHour:        lowHour,
		LowCount:       minVisitors,
		AveragePerHour: averagePerHour,
		TotalVisitors:  totalVisitors,
		PeakVsLowRatio: peakVsLowRatio,
	}

	insights := r.generatePeakHoursInsights(results, peakHour, lowHour)

	return &entity.PeakHoursAnalysis{
		Data:     dataPoints,
		Summary:  summary,
		Insights: insights,
	}, nil
}

func (r *PeopleCountRepositoryImpl) generatePeakHoursInsights(results []HourlyResult, peakHour, lowHour int) entity.PeakHoursInsights {
	busiestPeriod := ""
	quietestPeriod := ""

	switch {
	case peakHour >= 6 && peakHour < 12:
		busiestPeriod = "Morning (06:00-12:00)"
	case peakHour >= 12 && peakHour < 18:
		busiestPeriod = "Afternoon (12:00-18:00)"
	case peakHour >= 18 && peakHour < 24:
		busiestPeriod = "Evening (18:00-24:00)"
	default:
		busiestPeriod = "Late Night/Early Morning (00:00-06:00)"
	}

	switch {
	case lowHour >= 0 && lowHour < 6:
		quietestPeriod = "Late Night/Early Morning (00:00-06:00)"
	case lowHour >= 6 && lowHour < 12:
		quietestPeriod = "Morning (06:00-12:00)"
	case lowHour >= 12 && lowHour < 18:
		quietestPeriod = "Afternoon (12:00-18:00)"
	default:
		quietestPeriod = "Evening (18:00-24:00)"
	}

	totalVisitors := 0
	for _, result := range results {
		totalVisitors += result.Visitors
	}
	average := totalVisitors / len(results)

	var recommendedStaffingHours []string
	for _, result := range results {
		if result.Visitors >= average {
			recommendedStaffingHours = append(recommendedStaffingHours, fmt.Sprintf("%02d:00", result.Hour))
		}
	}

	trafficPattern := "Stable"
	if len(results) > 0 {
		max := 0
		min := results[0].Visitors
		for _, result := range results {
			if result.Visitors > max {
				max = result.Visitors
			}
			if result.Visitors < min {
				min = result.Visitors
			}
		}

		if min > 0 {
			ratio := float64(max) / float64(min)
			switch {
			case ratio >= 5.0:
				trafficPattern = "Highly Variable"
			case ratio >= 3.0:
				trafficPattern = "Variable"
			case ratio >= 2.0:
				trafficPattern = "Moderate"
			}
		}
	}

	return entity.PeakHoursInsights{
		BusiestPeriod:            busiestPeriod,
		QuietestPeriod:           quietestPeriod,
		RecommendedStaffingHours: recommendedStaffingHours,
		TrafficPattern:           trafficPattern,
	}
}
