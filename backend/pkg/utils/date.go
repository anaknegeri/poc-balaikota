package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeFormat defines standard time formats supported by the application
type TimeFormat string

const (
	// Standard formats
	RFC3339Format       TimeFormat = time.RFC3339
	RFC3339NanoFormat   TimeFormat = time.RFC3339Nano
	ISO8601Format       TimeFormat = "2006-01-02T15:04:05"
	ISO8601WithMsFormat TimeFormat = "2006-01-02T15:04:05.999999"

	// Custom formats
	DateTimeFormat TimeFormat = "2006-01-02 15:04:05"
	DateOnlyFormat TimeFormat = "2006-01-02"
	TimeOnlyFormat TimeFormat = "15:04:05"

	// Special formats for device timestamps
	DeviceBasicFormat        TimeFormat = "20060102_150405"
	DeviceWithMsFormat       TimeFormat = "20060102_150405_000000"
	DeviceYMDHMSFormat       TimeFormat = "20060102150405"
	DeviceYMDHMSWithMsFormat TimeFormat = "20060102150405.000"
)

// Standard format arrays for different parsing strategies
var (
	// AllFormats contains all supported time formats for comprehensive parsing
	AllFormats = []TimeFormat{
		RFC3339Format,
		RFC3339NanoFormat,
		ISO8601Format,
		ISO8601WithMsFormat,
		DateTimeFormat,
		DateOnlyFormat,
		TimeOnlyFormat,
		DeviceBasicFormat,
		DeviceWithMsFormat,
		DeviceYMDHMSFormat,
		DeviceYMDHMSWithMsFormat,
	}

	// CommonFormats contains frequently used formats
	CommonFormats = []TimeFormat{
		RFC3339Format,
		DateTimeFormat,
		DeviceBasicFormat,
	}

	// DeviceFormats focuses on device-specific timestamp formats
	DeviceFormats = []TimeFormat{
		DeviceBasicFormat,
		DeviceWithMsFormat,
		DeviceYMDHMSFormat,
		DeviceYMDHMSWithMsFormat,
	}
)

// ParseTime tries to parse a timestamp string using all supported formats
// Returns the parsed time and nil error if successful, or zero time and error if parsing fails
func ParseTime(timestamp string) (time.Time, error) {
	return ParseTimeWithFormats(timestamp, AllFormats)
}

// ParseTimeWithFormats tries to parse a timestamp using the specified formats
// Returns the parsed time and nil error if successful, or zero time and error if parsing fails
func ParseTimeWithFormats(timestamp string, formats []TimeFormat) (time.Time, error) {
	var parsedTime time.Time
	var err error

	// Try each format in order
	for _, format := range formats {
		parsedTime, err = time.Parse(string(format), timestamp)
		if err == nil {
			return parsedTime, nil
		}
	}

	// If standard formats fail, try specialized handling for complex formats
	parsedTime, err = parseComplexFormats(timestamp)
	if err == nil {
		return parsedTime, nil
	}

	return time.Time{}, fmt.Errorf("failed to parse timestamp '%s' with all known formats", timestamp)
}

// ParseDeviceTime is specialized for device timestamp formats
func ParseDeviceTime(timestamp string) (time.Time, error) {
	return ParseTimeWithFormats(timestamp, DeviceFormats)
}

// parseComplexFormats handles timestamps with special formats or extra components
func parseComplexFormats(timestamp string) (time.Time, error) {
	// Try to handle formats with milliseconds/microseconds as separate components
	if strings.Count(timestamp, "_") >= 2 {
		// Format like "20250518_221740_573760"
		parts := strings.Split(timestamp, "_")
		if len(parts) >= 2 {
			// Try to parse the base timestamp
			baseTimestamp := parts[0] + "_" + parts[1]
			parsedTime, err := time.Parse(string(DeviceBasicFormat), baseTimestamp)
			if err == nil && len(parts) > 2 {
				// If there's a third part, try to interpret as milliseconds/microseconds
				msStr := parts[2]
				if ms, err := strconv.ParseInt(msStr, 10, 64); err == nil {
					// Add the fractional seconds
					// Determine if it's milliseconds or microseconds based on length
					var duration time.Duration
					switch len(msStr) {
					case 3: // milliseconds
						duration = time.Duration(ms) * time.Millisecond
					case 6: // microseconds
						duration = time.Duration(ms) * time.Microsecond
					default: // assume nanoseconds or handle specifically
						if len(msStr) > 6 {
							// Truncate to microseconds if needed
							msStr = msStr[:6]
							ms, _ = strconv.ParseInt(msStr, 10, 64)
							duration = time.Duration(ms) * time.Microsecond
						} else {
							// Pad to microseconds
							factor := int64(1)
							for i := 0; i < (6 - len(msStr)); i++ {
								factor *= 10
							}
							duration = time.Duration(ms*factor) * time.Microsecond
						}
					}
					parsedTime = parsedTime.Add(duration)
				}
				return parsedTime, nil
			}
			return parsedTime, nil
		}
	}

	// Try to handle Unix timestamps (seconds since epoch)
	if unixSec, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		// If the timestamp is reasonable (between 2000 and 2100)
		if unixSec > 946684800 && unixSec < 4102444800 { // 2000-01-01 to 2100-01-01
			return time.Unix(unixSec, 0), nil
		}
	}

	// Try to handle Unix millisecond timestamps
	if len(timestamp) >= 13 {
		if unixMs, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
			// If reasonable millisecond timestamp (between 2000 and 2100)
			if unixMs > 946684800000 && unixMs < 4102444800000 {
				return time.Unix(unixMs/1000, (unixMs%1000)*1000000), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized timestamp format: %s", timestamp)
}

// FormatTime formats a time.Time according to the specified format
func FormatTime(t time.Time, format TimeFormat) string {
	return t.Format(string(format))
}

// ParseOrNow attempts to parse a timestamp, and returns current time if parsing fails
func ParseOrNow(timestamp string) time.Time {
	t, err := ParseTime(timestamp)
	if err != nil {
		return time.Now()
	}
	return t
}

// ParseOrDefault attempts to parse a timestamp, and returns a default time if parsing fails
func ParseOrDefault(timestamp string, defaultTime time.Time) time.Time {
	t, err := ParseTime(timestamp)
	if err != nil {
		return defaultTime
	}
	return t
}

// MustParse parses a timestamp and panics if parsing fails
// Only use when the format is guaranteed to be correct
func MustParse(timestamp string) time.Time {
	t, err := ParseTime(timestamp)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse timestamp: %s, error: %v", timestamp, err))
	}
	return t
}

// IsValidTimestamp checks if a string can be parsed as a timestamp
func IsValidTimestamp(timestamp string) bool {
	_, err := ParseTime(timestamp)
	return err == nil
}

// AgeInHours calculates the age of a timestamp in hours
func AgeInHours(t time.Time) float64 {
	return time.Since(t).Hours()
}

// AgeInDays calculates the age of a timestamp in days
func AgeInDays(t time.Time) float64 {
	return time.Since(t).Hours() / 24
}

func ParseAndValidateDate(dateStr, fieldName string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil // Empty date is allowed
	}

	// List of supported date formats
	formats := []string{
		time.RFC3339,                // 2025-05-13T10:00:00Z
		time.RFC3339Nano,            // 2025-05-13T10:00:00.123456789Z
		"2006-01-02T15:04:05",       // 2025-05-13T10:00:00
		"2006-01-02",                // 2025-05-13 (date only)
		"2006-01-02 15:04:05",       // 2025-05-13 10:00:00
		"2006-01-02T15:04:05-07:00", // 2025-05-13T10:00:00+07:00
	}

	var parsedTime time.Time
	var err error

	// Try each format
	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			// Successfully parsed
			if format == "2006-01-02" {
				// For date-only format, we need to decide if it's start or end of day
				if strings.Contains(fieldName, "to") || strings.Contains(fieldName, "end") {
					// For "to" dates, set to end of day
					parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 23, 59, 59, 999999999, parsedTime.Location())
				} else {
					// For "from" dates, set to start of day
					parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, parsedTime.Location())
				}
			}
			return parsedTime, nil
		}
	}

	// If all formats failed, return error with helpful message
	return time.Time{}, fmt.Errorf("invalid '%s' date format. Supported formats: YYYY-MM-DD, YYYY-MM-DDTHH:mm:ss, or RFC3339 (e.g. 2025-05-13T10:00:00Z)", fieldName)
}
