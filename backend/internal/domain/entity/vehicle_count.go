package entity

import (
	"time"
)

// VehicleCount model for vehicle_counts hypertable with TimescaleDB
type VehicleCount struct {
	ID                 string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CctvID             uint      `gorm:"not null;column:cctv_id" json:"cctv_id"`
	Timestamp          time.Time `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP;primaryKey;column:timestamp" json:"timestamp"`
	InCountCar         int       `gorm:"default:0;column:in_count_car" json:"in_count_car"`
	InCountTruck       int       `gorm:"default:0;column:in_count_truck" json:"in_count_truck"`
	InCountPeople      int       `gorm:"default:0;column:in_count_people" json:"in_count_people"`
	OutCount           int       `gorm:"default:0;column:out_count" json:"out_count"`
	DeviceTimestamp    time.Time `gorm:"type:timestamp with time zone;column:device_timestamp" json:"device_timestamp"`
	DeviceTimestampUTC float64   `gorm:"column:device_timestamp_utc" json:"device_timestamp_utc"`

	TotalInCount        int       `gorm:"->;column:total_in_count" json:"total_in_count"`
	TotalVehicleInCount int       `gorm:"->;column:total_vehicle_in_count" json:"total_vehicle_in_count"`
	CreatedAt           time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`

	// Relationships
	Cctv Camera `gorm:"foreignKey:CctvID" json:"cctv,omitempty"`
}

// TableName returns the table name for the VehicleCount model
func (VehicleCount) TableName() string {
	return "vehicle_counts"
}

// BeforeSave hook to calculate total counts before saving
func (v *VehicleCount) BeforeSave() error {
	// Calculate totals (won't be stored in DB if using GENERATED ALWAYS AS)
	v.TotalInCount = v.InCountCar + v.InCountTruck + v.InCountPeople
	v.TotalVehicleInCount = v.InCountCar + v.InCountTruck
	return nil
}

// VehicleTotalCounts represents aggregated vehicle counts across areas
type VehicleTotalCounts struct {
	InCar          int `json:"in_car"`
	InTruck        int `json:"in_truck"`
	InPeople       int `json:"in_people"`
	Out            int `json:"out"`
	TotalIn        int `json:"total_in"`
	TotalVehicleIn int `json:"total_vehicle_in"`
	NetCount       int `json:"net_count"` // Total In - Out
}

// VehicleCountSummary represents the current vehicle count summary across all areas
type VehicleCountSummary struct {
	Cctvs  []CctvVehicleSummary `json:"cctvs"`
	Totals VehicleTotalCounts   `json:"totals"`
}

type CctvVehicleSummary struct {
	CctvID              uint      `json:"cctv_id"`
	CctvName            string    `json:"cctv_name"`
	InCountCar          int       `json:"in_count_car"`
	InCountTruck        int       `json:"in_count_truck"`
	InCountPeople       int       `json:"in_count_people"`
	OutCount            int       `json:"out_count"`
	TotalInCount        int       `json:"total_in_count"`
	TotalVehicleInCount int       `json:"total_vehicle_in_count"`
	NetCount            int       `json:"net_count"`
	LastUpdated         time.Time `json:"last_updated"`
}

// VehicleTrendPoint represents a data point in vehicle trend analysis
type VehicleTrendPoint struct {
	TimePeriod          time.Time `json:"time_period"`
	InCountCar          int       `json:"in_count_car"`
	InCountTruck        int       `json:"in_count_truck"`
	InCountPeople       int       `json:"in_count_people"`
	OutCount            int       `json:"out_count"`
	TotalInCount        int       `json:"total_in_count"`
	TotalVehicleInCount int       `json:"total_vehicle_in_count"`
	NetCount            int       `json:"net_count"`
}

// VehicleCountsByTimeResult is a time-based grouping of vehicle counts
type VehicleCountsByTimeResult struct {
	Interval string              `json:"interval"`
	Data     []VehicleTrendPoint `json:"data"`
}

// CalculateTotalCounts calculates all total counts for the vehicle record
func (v *VehicleCount) CalculateTotalCounts() {
	v.TotalInCount = v.InCountCar + v.InCountTruck + v.InCountPeople
	v.TotalVehicleInCount = v.InCountCar + v.InCountTruck
}

// GetNetCount returns the net count (total in - out)
func (v *VehicleCount) GetNetCount() int {
	return v.TotalInCount - v.OutCount
}

// IsVehicleOnly checks if the count only includes vehicles (no people)
func (v *VehicleCount) IsVehicleOnly() bool {
	return v.InCountPeople == 0
}

// HasOutboundData checks if there's any outbound count data
func (v *VehicleCount) HasOutboundData() bool {
	return v.OutCount > 0
}
