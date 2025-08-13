package utils

import (
	"path/filepath"
)

// ConvertFilePathToURL converts a local file path to a web-accessible URL
// baseURL should be the base URL of your application, e.g., "http://localhost:8080"
func ConvertFilePathToURL(filePath, baseURL, staticRoute string) string {
	if filePath == "" {
		return ""
	}

	// Extract just the filename from the path
	filename := filepath.Base(filePath)

	// Combine with the base URL and route
	return baseURL + staticRoute + "/" + filename
}

// ConvertFaceImagePathToURL specifically converts face recognition image paths
func ConvertFaceImagePathToURL(filePath, baseURL string) string {
	return ConvertFilePathToURL(filePath, baseURL, "/api/images/faces")
}
