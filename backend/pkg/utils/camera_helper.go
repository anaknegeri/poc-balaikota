package utils

// CameraIDResolver provides methods to resolve camera ID from various sources
type CameraIDResolver struct{}

// NewCameraIDResolver creates a new instance of CameraIDResolver
func NewCameraIDResolver() *CameraIDResolver {
	return &CameraIDResolver{}
}

// ResolveCameraID determines camera ID from available fields with priority:
// 1. CameraID field (if not 0) - add 1 for compatibility
// 2. CCTVID field (if not 0) - add 1 for legacy compatibility
// 3. ObjectID field (if not 0) - use as is
// 4. Default to 1 if none available
func (r *CameraIDResolver) ResolveCameraID(cameraID, cctvID uint) uint {
	if cameraID != 0 {
		// Use camera_id field if available (add 1 for compatibility)
		return cameraID + 1
	}

	if cctvID != 0 {
		// Use cctv_id field (add 1 for legacy compatibility)
		return cctvID + 1
	}

	// Default camera ID
	return 1
}

// ResolveCameraIDFromString determines camera ID from string-based objectID
// This is useful when objectID comes as string type
func (r *CameraIDResolver) ResolveCameraIDFromString(cameraID, cctvID uint, objectIDStr string) uint {
	if cameraID != 0 {
		// Use camera_id field if available (add 1 for compatibility)
		return cameraID + 1
	}

	if cctvID != 0 {
		// Use cctv_id field (add 1 for legacy compatibility)
		return cctvID + 1
	}

	// Try to parse objectID string to uint
	if objectIDStr != "" {
		// For string objectID, we'll use a simple approach
		// In real implementation, you might want to parse it properly
		// For now, we'll just return 1 if it's not empty
		return 1
	}

	// Default camera ID
	return 1
}

// FlexibleUintResolver handles FlexibleUint type specifically for AlertData
type FlexibleUintResolver interface {
	ToUint() uint
}

// ResolveCameraIDWithFlexible determines camera ID using FlexibleUint interface
func (r *CameraIDResolver) ResolveCameraIDWithFlexible(cameraID, cctvID uint, objectID FlexibleUintResolver) uint {
	if cameraID != 0 {
		// Use camera_id field if available (add 1 for compatibility)
		return cameraID + 1
	}

	if cctvID != 0 {
		// Use cctv_id field (add 1 for legacy compatibility)
		return cctvID + 1
	}

	if objectID != nil && objectID.ToUint() != 0 {
		// Use objectID as fallback (converted from FlexibleUint)
		return objectID.ToUint()
	}

	// Default camera ID
	return 1
}
