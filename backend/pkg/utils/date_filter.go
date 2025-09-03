package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// DateRange represents a date range filter
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// DateFilterOptions configures date filter behavior
type DateFilterOptions struct {
	// DefaultToEndOfDay sets whether "to" dates should default to end of day (23:59:59)
	DefaultToEndOfDay bool
	// DefaultFromStartOfDay sets whether "from" dates should default to start of day (00:00:00)
	DefaultFromStartOfDay bool
	// AllowEmptyDates allows empty from/to parameters
	AllowEmptyDates bool
	// MaxRangeDays limits the maximum allowed date range in days (0 = no limit)
	MaxRangeDays int
}

// DefaultDateFilterOptions returns sensible defaults
func DefaultDateFilterOptions() DateFilterOptions {
	return DateFilterOptions{
		DefaultToEndOfDay:     true,
		DefaultFromStartOfDay: true,
		AllowEmptyDates:       true,
		MaxRangeDays:          0, // No limit
	}
}

// parseFlexibleDate parses various date formats
func parseFlexibleDate(dateStr, fieldName string, isEndOfDay bool) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil // Empty date is allowed
	}

	// Load Asia/Jakarta timezone for Indonesian applications
	jakartaTZ, _ := time.LoadLocation("Asia/Jakarta")
	if jakartaTZ == nil {
		// Fallback to UTC+7 if timezone loading fails
		jakartaTZ = time.FixedZone("WIB", 7*60*60)
	}

	// List of supported date formats
	formats := []string{
		time.RFC3339,                // 2025-05-13T10:00:00Z
		time.RFC3339Nano,            // 2025-05-13T10:00:00.123456789Z
		"2006-01-02T15:04:05",       // 2025-05-13T10:00:00
		"2006-01-02",                // 2025-05-13 (date only)
		"2006-01-02 15:04:05",       // 2025-05-13 10:00:00
		"2006-01-02T15:04:05-07:00", // 2025-05-13T10:00:00+07:00
		"2006-01-02T15:04:05.000Z",  // 2025-05-13T10:00:00.000Z
		"20060102_150405",           // 20250513_100000 (compact format)
		"20060102_150405_000000",    // 20250513_100000_123456 (with microseconds)
		"20060102",                  // 20250513 (compact date only)
		"02/01/2006",                // DD/MM/YYYY
		"01/02/2006",                // MM/DD/YYYY
		"2006/01/02",                // YYYY/MM/DD
		"02-01-2006",                // DD-MM-YYYY
		"01-02-2006",                // MM-DD-YYYY
		"2006-01-02 15:04:05.000",   // 2025-05-13 10:00:00.000
		"2006-01-02T15:04:05.000",   // 2025-05-13T10:00:00.000
	}

	var parsedTime time.Time
	var err error

	// Try each format
	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			// Successfully parsed
			if format == "2006-01-02" || format == "20060102" || strings.Contains(format, "/") {
				// For date-only formats, adjust time based on parameters and use Jakarta timezone
				if isEndOfDay {
					// Set to end of day in Jakarta timezone
					parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 23, 59, 59, 999999999, jakartaTZ)
				} else {
					// Set to start of day in Jakarta timezone
					parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, jakartaTZ)
				}
			} else if parsedTime.Location() == time.UTC && !strings.Contains(dateStr, "Z") && !strings.Contains(dateStr, "+") && !strings.Contains(dateStr, "-") {
				// If parsed as UTC but no timezone info in string, convert to Jakarta timezone
				parsedTime = parsedTime.In(jakartaTZ)
			}
			return parsedTime, nil
		}
	}

	// If all formats failed, try to handle special cases with partial parsing
	return parseFlexibleDateFallback(dateStr, fieldName, isEndOfDay)
}

// parseFlexibleDateFallback handles special cases and partial date parsing
func parseFlexibleDateFallback(dateStr, fieldName string, isEndOfDay bool) (time.Time, error) {
	// Load Asia/Jakarta timezone for Indonesian applications
	jakartaTZ, _ := time.LoadLocation("Asia/Jakarta")
	if jakartaTZ == nil {
		// Fallback to UTC+7 if timezone loading fails
		jakartaTZ = time.FixedZone("WIB", 7*60*60)
	}

	// Handle timestamps with extra microseconds (like: 20250730_113717_045594)
	if len(dateStr) > 15 && strings.Contains(dateStr, "_") {
		parts := strings.Split(dateStr, "_")
		if len(parts) >= 2 {
			// Try to extract date and time parts
			dateTimePart := parts[0] + "_" + parts[1]

			// Try standard format first
			parsedTime, err := time.Parse("20060102_150405", dateTimePart)
			if err == nil {
				return parsedTime.In(jakartaTZ), nil
			}

			// If that fails, try just the date part
			if len(parts[0]) == 8 {
				parsedTime, err := time.Parse("20060102", parts[0])
				if err == nil {
					if isEndOfDay {
						parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 23, 59, 59, 999999999, jakartaTZ)
					} else {
						parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, jakartaTZ)
					}
					return parsedTime, nil
				}
			}
		}
	}

	// Handle ISO-like formats with milliseconds
	if strings.Contains(dateStr, "T") && (strings.Contains(dateStr, ".") || strings.Contains(dateStr, "Z")) {
		// Try to truncate milliseconds if too long
		if idx := strings.Index(dateStr, "."); idx != -1 {
			beforeDot := dateStr[:idx]
			afterDot := dateStr[idx+1:]

			// Find end of milliseconds (before Z or timezone)
			endIdx := len(afterDot)
			for i, ch := range afterDot {
				if ch == 'Z' || ch == '+' || ch == '-' {
					endIdx = i
					break
				}
			}

			// Truncate to 6 digits (microseconds) or less
			msec := afterDot[:endIdx]
			if len(msec) > 6 {
				msec = msec[:6]
			}

			remainder := afterDot[endIdx:]
			truncated := beforeDot + "." + msec + remainder

			formats := []string{
				time.RFC3339Nano,
				"2006-01-02T15:04:05.000000Z",
				"2006-01-02T15:04:05.000Z",
				"2006-01-02T15:04:05.000000",
				"2006-01-02T15:04:05.000",
			}

			for _, format := range formats {
				if parsedTime, err := time.Parse(format, truncated); err == nil {
					return parsedTime, nil
				}
			}
		}
	}

	// Handle relative dates (today, yesterday, etc.)
	lower := strings.ToLower(strings.TrimSpace(dateStr))
	// Load Asia/Jakarta timezone for Indonesian applications
	jakartaTZ2, _ := time.LoadLocation("Asia/Jakarta")
	if jakartaTZ2 == nil {
		// Fallback to UTC+7 if timezone loading fails
		jakartaTZ2 = time.FixedZone("WIB", 7*60*60)
	}
	now := time.Now().In(jakartaTZ2)

	switch lower {
	case "today", "now":
		if isEndOfDay {
			return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, jakartaTZ2), nil
		}
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jakartaTZ2), nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		if isEndOfDay {
			return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, jakartaTZ2), nil
		}
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, jakartaTZ2), nil
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1)
		if isEndOfDay {
			return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 999999999, jakartaTZ2), nil
		}
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, jakartaTZ2), nil
	}

	// If all parsing attempts failed, return error with helpful message
	return time.Time{}, fmt.Errorf("invalid '%s' date format '%s'. Supported formats: YYYY-MM-DD, YYYY-MM-DDTHH:mm:ss, RFC3339, DD/MM/YYYY, YYYYMMDD_HHMMSS, or relative dates (today, yesterday, tomorrow)", fieldName, dateStr)
}

// ParseDateRangeFromQuery extracts and validates date range from Fiber context query parameters
func ParseDateRangeFromQuery(c *fiber.Ctx, options ...DateFilterOptions) (*DateRange, error) {
	opts := DefaultDateFilterOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")

	return ParseDateRangeFromStrings(fromStr, toStr, opts)
}

// ParseDateRangeFromStrings parses date range from string parameters
func ParseDateRangeFromStrings(fromStr, toStr string, options ...DateFilterOptions) (*DateRange, error) {
	opts := DefaultDateFilterOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	dateRange := &DateRange{}

	// Parse from date
	if fromStr != "" {
		fromTime, err := parseFlexibleDate(fromStr, "from", !opts.DefaultFromStartOfDay)
		if err != nil {
			return nil, err
		}
		if !fromTime.IsZero() {
			dateRange.From = &fromTime
		}
	} else if !opts.AllowEmptyDates {
		return nil, fmt.Errorf("'from' date is required")
	}

	// Parse to date
	if toStr != "" {
		toTime, err := parseFlexibleDate(toStr, "to", opts.DefaultToEndOfDay)
		if err != nil {
			return nil, err
		}
		if !toTime.IsZero() {
			dateRange.To = &toTime
		}
	} else if !opts.AllowEmptyDates {
		return nil, fmt.Errorf("'to' date is required")
	}

	// Validate date range logic
	if dateRange.From != nil && dateRange.To != nil {
		if dateRange.From.After(*dateRange.To) {
			return nil, fmt.Errorf("invalid date range: 'from' date must be before or equal to 'to' date")
		}

		// Check maximum range if specified
		if opts.MaxRangeDays > 0 {
			daysDiff := int(dateRange.To.Sub(*dateRange.From).Hours() / 24)
			if daysDiff > opts.MaxRangeDays {
				return nil, fmt.Errorf("date range too large: maximum allowed range is %d days", opts.MaxRangeDays)
			}
		}
	}

	return dateRange, nil
}

// ToRFC3339Strings converts DateRange to RFC3339 formatted strings
func (dr *DateRange) ToRFC3339Strings() (from, to string) {
	if dr.From != nil {
		from = dr.From.Format(time.RFC3339)
	}
	if dr.To != nil {
		to = dr.To.Format(time.RFC3339)
	}
	return from, to
}

// ToSQLStrings converts DateRange to SQL datetime formatted strings
func (dr *DateRange) ToSQLStrings() (from, to string) {
	if dr.From != nil {
		from = dr.From.Format("2006-01-02 15:04:05")
	}
	if dr.To != nil {
		to = dr.To.Format("2006-01-02 15:04:05")
	}
	return from, to
}

// ToDateOnlyStrings converts DateRange to date-only formatted strings
func (dr *DateRange) ToDateOnlyStrings() (from, to string) {
	if dr.From != nil {
		from = dr.From.Format("2006-01-02")
	}
	if dr.To != nil {
		to = dr.To.Format("2006-01-02")
	}
	return from, to
}

// IsEmpty checks if the date range is empty (both from and to are nil)
func (dr *DateRange) IsEmpty() bool {
	return dr.From == nil && dr.To == nil
}

// HasFrom checks if from date is set
func (dr *DateRange) HasFrom() bool {
	return dr.From != nil
}

// HasTo checks if to date is set
func (dr *DateRange) HasTo() bool {
	return dr.To != nil
}

// DaysDifference returns the number of days between from and to dates
func (dr *DateRange) DaysDifference() int {
	if dr.From == nil || dr.To == nil {
		return 0
	}
	return int(dr.To.Sub(*dr.From).Hours() / 24)
}

// ApplyToQuery applies date range to SQL query conditions
// This is a helper for building WHERE clauses
func (dr *DateRange) ApplyToQuery(fieldName string, conditions []string, args []interface{}) ([]string, []interface{}) {
	if dr.From != nil {
		conditions = append(conditions, fmt.Sprintf("%s >= ?", fieldName))
		args = append(args, *dr.From)
	}
	if dr.To != nil {
		conditions = append(conditions, fmt.Sprintf("%s <= ?", fieldName))
		args = append(args, *dr.To)
	}
	return conditions, args
}

// HandleDateFilterError creates a standardized error response for date filter errors
func HandleDateFilterError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": true,
		"msg":   err.Error(),
	})
}

// FlexibleDateParser is a more advanced date parser with configuration options
type FlexibleDateParser struct {
	// CustomFormats allows adding custom date formats
	CustomFormats []string
	// DefaultTimezone sets the default timezone for dates without timezone info
	DefaultTimezone *time.Location
	// StrictMode when true, only allows predefined formats
	StrictMode bool
	// AllowRelativeDates enables parsing of relative dates like "today", "yesterday"
	AllowRelativeDates bool
}

// NewFlexibleDateParser creates a new parser with default settings
func NewFlexibleDateParser() *FlexibleDateParser {
	return &FlexibleDateParser{
		CustomFormats:      []string{},
		DefaultTimezone:    time.Local,
		StrictMode:         false,
		AllowRelativeDates: true,
	}
}

// AddCustomFormat adds a custom date format to the parser
func (p *FlexibleDateParser) AddCustomFormat(format string) {
	p.CustomFormats = append(p.CustomFormats, format)
}

// SetTimezone sets the default timezone for parsing
func (p *FlexibleDateParser) SetTimezone(timezone *time.Location) {
	p.DefaultTimezone = timezone
}

// Parse parses a date string using the flexible parser
func (p *FlexibleDateParser) Parse(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	// Try custom formats first
	for _, format := range p.CustomFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// Handle relative dates if enabled
	if p.AllowRelativeDates {
		if t, ok := p.parseRelativeDate(dateStr); ok {
			return t, nil
		}
	}

	// Use the existing parseFlexibleDate function
	return parseFlexibleDate(dateStr, "date", false)
}

// parseRelativeDate handles relative date parsing
func (p *FlexibleDateParser) parseRelativeDate(dateStr string) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(dateStr))
	now := time.Now()

	if p.DefaultTimezone != nil {
		now = now.In(p.DefaultTimezone)
	}

	switch {
	case lower == "now" || lower == "today":
		return now, true
	case lower == "yesterday":
		return now.AddDate(0, 0, -1), true
	case lower == "tomorrow":
		return now.AddDate(0, 0, 1), true
	case strings.HasSuffix(lower, " ago"):
		return p.parseRelativePast(lower, now)
	case strings.HasPrefix(lower, "in "):
		return p.parseRelativeFuture(lower, now)
	case strings.HasPrefix(lower, "last "):
		return p.parseLastPeriod(lower, now)
	case strings.HasPrefix(lower, "next "):
		return p.parseNextPeriod(lower, now)
	}

	return time.Time{}, false
}

// parseRelativePast handles "X ago" format
func (p *FlexibleDateParser) parseRelativePast(dateStr string, now time.Time) (time.Time, bool) {
	// Remove " ago" suffix
	timeStr := strings.TrimSuffix(dateStr, " ago")

	// Parse patterns like "1 day", "2 weeks", "3 months", "1 year"
	parts := strings.Fields(timeStr)
	if len(parts) != 2 {
		return time.Time{}, false
	}

	amount := 0
	if _, err := fmt.Sscanf(parts[0], "%d", &amount); err != nil {
		return time.Time{}, false
	}

	unit := parts[1]
	switch {
	case strings.HasPrefix(unit, "second"):
		return now.Add(-time.Duration(amount) * time.Second), true
	case strings.HasPrefix(unit, "minute"):
		return now.Add(-time.Duration(amount) * time.Minute), true
	case strings.HasPrefix(unit, "hour"):
		return now.Add(-time.Duration(amount) * time.Hour), true
	case strings.HasPrefix(unit, "day"):
		return now.AddDate(0, 0, -amount), true
	case strings.HasPrefix(unit, "week"):
		return now.AddDate(0, 0, -amount*7), true
	case strings.HasPrefix(unit, "month"):
		return now.AddDate(0, -amount, 0), true
	case strings.HasPrefix(unit, "year"):
		return now.AddDate(-amount, 0, 0), true
	}

	return time.Time{}, false
}

// parseRelativeFuture handles "in X" format
func (p *FlexibleDateParser) parseRelativeFuture(dateStr string, now time.Time) (time.Time, bool) {
	// Remove "in " prefix
	timeStr := strings.TrimPrefix(dateStr, "in ")

	parts := strings.Fields(timeStr)
	if len(parts) != 2 {
		return time.Time{}, false
	}

	amount := 0
	if _, err := fmt.Sscanf(parts[0], "%d", &amount); err != nil {
		return time.Time{}, false
	}

	unit := parts[1]
	switch {
	case strings.HasPrefix(unit, "second"):
		return now.Add(time.Duration(amount) * time.Second), true
	case strings.HasPrefix(unit, "minute"):
		return now.Add(time.Duration(amount) * time.Minute), true
	case strings.HasPrefix(unit, "hour"):
		return now.Add(time.Duration(amount) * time.Hour), true
	case strings.HasPrefix(unit, "day"):
		return now.AddDate(0, 0, amount), true
	case strings.HasPrefix(unit, "week"):
		return now.AddDate(0, 0, amount*7), true
	case strings.HasPrefix(unit, "month"):
		return now.AddDate(0, amount, 0), true
	case strings.HasPrefix(unit, "year"):
		return now.AddDate(amount, 0, 0), true
	}

	return time.Time{}, false
}

// parseLastPeriod handles "last X" format
func (p *FlexibleDateParser) parseLastPeriod(dateStr string, now time.Time) (time.Time, bool) {
	period := strings.TrimPrefix(dateStr, "last ")

	switch period {
	case "week":
		// Go to start of last week (Monday)
		daysToSubtract := int(now.Weekday())
		if daysToSubtract == 0 {
			daysToSubtract = 7 // Sunday
		}
		daysToSubtract += 6 // Go to previous week
		return now.AddDate(0, 0, -daysToSubtract), true
	case "month":
		// Go to first day of last month
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return firstOfThisMonth.AddDate(0, -1, 0), true
	case "year":
		// Go to first day of last year
		return time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location()), true
	}

	return time.Time{}, false
}

// parseNextPeriod handles "next X" format
func (p *FlexibleDateParser) parseNextPeriod(dateStr string, now time.Time) (time.Time, bool) {
	period := strings.TrimPrefix(dateStr, "next ")

	switch period {
	case "week":
		// Go to start of next week (Monday)
		daysToAdd := 7 - int(now.Weekday())
		if now.Weekday() == time.Sunday {
			daysToAdd = 1
		} else {
			daysToAdd++
		}
		return now.AddDate(0, 0, daysToAdd), true
	case "month":
		// Go to first day of next month
		firstOfNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		return firstOfNextMonth, true
	case "year":
		// Go to first day of next year
		return time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location()), true
	}

	return time.Time{}, false
}

// ParseDateRange parses a date range string like "2025-01-01 to 2025-01-31" or "last week to today"
func (p *FlexibleDateParser) ParseDateRange(rangeStr string) (*DateRange, error) {
	// Split by common separators
	separators := []string{" to ", " - ", " ~ ", ".."}
	var parts []string

	for _, sep := range separators {
		if strings.Contains(rangeStr, sep) {
			parts = strings.Split(rangeStr, sep)
			break
		}
	}

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid date range format. Expected format: 'start to end' or 'start - end'")
	}

	fromTime, err := p.Parse(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %v", err)
	}

	toTime, err := p.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %v", err)
	}

	return &DateRange{
		From: &fromTime,
		To:   &toTime,
	}, nil
}

// SmartDateParser is a convenience function that uses FlexibleDateParser with smart defaults
func SmartDateParser(dateStr string) (time.Time, error) {
	parser := NewFlexibleDateParser()

	// Add some common custom formats
	parser.AddCustomFormat("2006-01-02 15:04")
	parser.AddCustomFormat("02/01/2006 15:04")
	parser.AddCustomFormat("Jan 2, 2006")
	parser.AddCustomFormat("January 2, 2006")
	parser.AddCustomFormat("2 Jan 2006")
	parser.AddCustomFormat("2 January 2006")

	return parser.Parse(dateStr)
}

// ParseDateRangeString is a convenience function for parsing date ranges
func ParseDateRangeString(rangeStr string) (*DateRange, error) {
	parser := NewFlexibleDateParser()
	return parser.ParseDateRange(rangeStr)
}

// Example usage in different scenarios:

// For handlers that need basic date filtering
func ExampleBasicUsage(c *fiber.Ctx) error {
	dateRange, err := ParseDateRangeFromQuery(c)
	if err != nil {
		return HandleDateFilterError(c, err)
	}

	from, to := dateRange.ToRFC3339Strings()
	// Use from and to in your service calls
	_ = from
	_ = to

	return c.JSON(fiber.Map{"success": true})
}

// For advanced date parsing with custom requirements
func ExampleAdvancedUsage() {
	parser := NewFlexibleDateParser()

	// Add custom formats for your specific use case
	parser.AddCustomFormat("20060102_150405")        // YYYYMMDD_HHMMSS
	parser.AddCustomFormat("20060102_150405_000000") // YYYYMMDD_HHMMSS_microseconds

	// Set timezone if needed
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	parser.SetTimezone(jakarta)

	// Parse various formats
	examples := []string{
		"2025-08-04",
		"today",
		"yesterday",
		"1 week ago",
		"in 3 days",
		"last month",
		"next year",
		"20250804_143000",
		"20250804_143000_123456",
		"Jan 1, 2025",
		"1/1/2025",
	}

	for _, example := range examples {
		if date, err := parser.Parse(example); err == nil {
			fmt.Printf("Parsed '%s' -> %s\n", example, date.Format(time.RFC3339))
		}
	}

	// Parse date ranges
	rangeExamples := []string{
		"yesterday to today",
		"2025-01-01 to 2025-12-31",
		"last week to next week",
		"1 month ago - today",
	}

	for _, example := range rangeExamples {
		if dateRange, err := parser.ParseDateRange(example); err == nil {
			from, to := dateRange.ToRFC3339Strings()
			fmt.Printf("Range '%s' -> %s to %s\n", example, from, to)
		}
	}
}

// For simple one-off parsing
func ExampleSimpleUsage() {
	// Use the convenience function for quick parsing
	date, err := SmartDateParser("yesterday")
	if err == nil {
		fmt.Printf("Yesterday was: %s\n", date.Format("2006-01-02"))
	}

	// Parse date range with convenience function
	dateRange, err := ParseDateRangeString("last week to today")
	if err == nil {
		fmt.Printf("Date range: %s to %s\n",
			dateRange.From.Format("2006-01-02"),
			dateRange.To.Format("2006-01-02"))
	}
}
