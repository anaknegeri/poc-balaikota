package entity

import (
	"time"
)

// PeopleCount model for people_counts hypertable with TimescaleDB
type PeopleCount struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CameraID     uint      `gorm:"not null;column:camera_id" json:"camera_id"`
	Timestamp    time.Time `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP;primaryKey;column:timestamp" json:"timestamp"`
	MaleCount    int       `gorm:"default:0;column:male_count" json:"male_count"`
	FemaleCount  int       `gorm:"default:0;column:female_count" json:"female_count"`
	ChildCount   int       `gorm:"default:0;column:child_count" json:"child_count"`
	AdultCount   int       `gorm:"default:0;column:adult_count" json:"adult_count"`
	ElderlyCount int       `gorm:"default:0;column:elderly_count" json:"elderly_count"`
	TotalCount   int       `gorm:"->;column:total_count" json:"total_count"`
	CreatedAt    time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`

	// Relationships
	Camera Camera `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
}

// TableName returns the table name for the PeopleCount model
func (PeopleCount) TableName() string {
	return "people_counts"
}

// BeforeSave hook to calculate TotalCount before saving
func (p *PeopleCount) BeforeSave() error {
	// Note: This won't actually be stored in DB as TotalCount is GENERATED ALWAYS AS
	// but can be used on the Go side before the record is saved
	p.TotalCount = p.MaleCount + p.FemaleCount
	return nil
}

// TotalCounts represents aggregated counts across areas
type TotalCounts struct {
	Male    int `json:"male"`
	Female  int `json:"female"`
	Child   int `json:"child"`
	Adult   int `json:"adult"`
	Elderly int `json:"elderly"`
	Total   int `json:"total"`
}

// CountSummary represents the current count summary across all areas
type CountSummary struct {
	Cameras []CameraSummary `json:"cameras"`
	Totals  TotalCounts     `json:"totals"`
}

type CameraSummary struct {
	CameraID     uint      `json:"camera_id"`
	CameraName   string    `json:"camera_name"`
	MaleCount    int       `json:"male_count"`
	FemaleCount  int       `json:"female_count"`
	ChildCount   int       `json:"child_count"`
	AdultCount   int       `json:"adult_count"`
	ElderlyCount int       `json:"elderly_count"`
	TotalCount   int       `json:"total_count"`
	LastUpdated  time.Time `json:"last_updated"`
}

// TrendPoint represents a data point in a trend analysis
type TrendPoint struct {
	TimePeriod   time.Time `json:"time_period"`
	MaleCount    int       `json:"male_count"`
	FemaleCount  int       `json:"female_count"`
	TotalCount   int       `json:"total_count"`
	ChildCount   int       `json:"child_count"`
	AdultCount   int       `json:"adult_count"`
	ElderlyCount int       `json:"elderly_count"`
}

// CountsByTimeResult is a time-based grouping of counts
type CountsByTimeResult struct {
	Interval string       `json:"interval"`
	Data     []TrendPoint `json:"data"`
}

// PeakHourPoint represents a data point for peak hours analysis
type PeakHourPoint struct {
	Hour       int    `json:"hour"`
	TimeLabel  string `json:"time_label"`
	Visitors   int    `json:"visitors"`
	Period     string `json:"period"` // peak, high, moderate, low
	Percentage int    `json:"percentage"`
}

// PeakHoursAnalysis represents the complete peak hours analysis
type PeakHoursAnalysis struct {
	Data     []PeakHourPoint   `json:"data"`
	Summary  PeakHoursSummary  `json:"summary"`
	Insights PeakHoursInsights `json:"insights"`
}

// PeakHoursSummary contains summary statistics
type PeakHoursSummary struct {
	PeakHour       int     `json:"peak_hour"`
	PeakCount      int     `json:"peak_count"`
	LowHour        int     `json:"low_hour"`
	LowCount       int     `json:"low_count"`
	AveragePerHour float64 `json:"average_per_hour"`
	TotalVisitors  int     `json:"total_visitors"`
	PeakVsLowRatio float64 `json:"peak_vs_low_ratio"`
}

// PeakHoursInsights contains business insights
type PeakHoursInsights struct {
	BusiestPeriod            string   `json:"busiest_period"`
	QuietestPeriod           string   `json:"quietest_period"`
	RecommendedStaffingHours []string `json:"recommended_staffing_hours"`
	TrafficPattern           string   `json:"traffic_pattern"`
}

func (pc *PeopleCount) CalculateTotalCount() {
	pc.TotalCount = pc.MaleCount + pc.FemaleCount
}
