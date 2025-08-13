package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"people-counting/internal/domain/service"

	"github.com/gofiber/fiber/v2"
)

// CameraStreamService provides MJPEG streaming capabilities for cameras
type CameraStreamService struct {
	cameraService      service.CameraService
	streams            map[uint]*CameraStream
	mu                 sync.Mutex
	baseImageDirectory string
	running            bool
	ctx                context.Context
	cancelFunc         context.CancelFunc
}

// CameraStream represents a single camera stream
type CameraStream struct {
	cameraID     uint
	clients      map[chan []byte]bool
	clientsMu    sync.Mutex
	lastActivity map[chan []byte]time.Time
	stopChan     chan struct{}
	frameCache   []byte
	frameRate    int
	quality      int
	imagePath    string
	isRunning    bool
	ctx          context.Context
	cancelFn     context.CancelFunc
}

// StreamConfig defines configuration for a stream
type StreamConfig struct {
	FrameRate int // frames per second
	Quality   int // JPEG quality (1-100)
}

// NewCameraStreamService creates a new camera streaming service
func NewCameraStreamService(cameraService service.CameraService, baseImageDirectory string) *CameraStreamService {
	ctx, cancel := context.WithCancel(context.Background())

	return &CameraStreamService{
		cameraService:      cameraService,
		streams:            make(map[uint]*CameraStream),
		baseImageDirectory: baseImageDirectory,
		ctx:                ctx,
		cancelFunc:         cancel,
	}
}

// RegisterRoutes registers the streaming routes to the Fiber router
func (s *CameraStreamService) RegisterRoutes(router fiber.Router) {
	// Stream endpoint for all cameras
	router.Get("/cameras/:id/stream", s.handleStream)

	// Static image endpoint for all cameras
	router.Get("/cameras/:id/image", s.handleImage)
}

// Start initializes the streaming service
func (s *CameraStreamService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil // Already running
	}

	// Ensure base directory exists
	if err := os.MkdirAll(s.baseImageDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create base image directory: %w", err)
	}

	s.running = true
	log.Println("Camera streaming service started")
	return nil
}

// Stop shuts down the streaming service
func (s *CameraStreamService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Stop all active streams
	for _, stream := range s.streams {
		stream.stop()
	}

	s.cancelFunc()
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancelFunc = cancel

	s.running = false
	log.Println("Camera streaming service stopped")
	return nil
}

// StreamExists checks if a stream exists for the given camera ID
func (s *CameraStreamService) StreamExists(cameraID uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream, exists := s.streams[cameraID]
	return exists && stream.isRunning
}

// GetCameraStreamURL returns the URL for a camera's stream
func (s *CameraStreamService) GetCameraStreamURL(c *fiber.Ctx, cameraID uint) string {
	baseURL := c.Protocol() + "://" + c.Hostname()
	return fmt.Sprintf("%s/api/cameras/%d/stream", baseURL, cameraID)
}

// GetCameraImageURL returns the URL for a camera's static image
func (s *CameraStreamService) GetCameraImageURL(c *fiber.Ctx, cameraID uint) string {
	baseURL := c.Protocol() + "://" + c.Hostname()
	return fmt.Sprintf("%s/api/cameras/%d/image", baseURL, cameraID)
}

// GetAllCameraStreamURLs returns the URLs for all active camera streams
func (s *CameraStreamService) GetAllCameraStreamURLs(c *fiber.Ctx) map[uint]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	baseURL := c.Protocol() + "://" + c.Hostname()
	urls := make(map[uint]string)
	for cameraID, stream := range s.streams {
		if stream.isRunning {
			urls[cameraID] = fmt.Sprintf("%s/api/cameras/%d/stream", baseURL, cameraID)
		}
	}

	return urls
}

// StartCameraStream starts a stream for a specific camera
func (s *CameraStreamService) StartCameraStream(c *fiber.Ctx, cameraID uint, config StreamConfig) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("streaming service is not running")
	}

	// Check if camera exists
	_, err := s.cameraService.GetCameraByID(c.Context(), cameraID)
	if err != nil {
		return "", fmt.Errorf("camera not found: %w", err)
	}

	// Check if stream is already active
	if stream, exists := s.streams[cameraID]; exists && stream.isRunning {
		return s.GetCameraStreamURL(c, cameraID), nil
	}

	// Create image path
	imagePath := filepath.Join(s.baseImageDirectory, fmt.Sprintf("image_%d.jpg", cameraID))

	// Create new stream
	ctx, cancel := context.WithCancel(s.ctx)
	stream := &CameraStream{
		cameraID:     cameraID,
		frameRate:    config.FrameRate,
		quality:      config.Quality,
		imagePath:    imagePath,
		clients:      make(map[chan []byte]bool),
		lastActivity: make(map[chan []byte]time.Time),
		stopChan:     make(chan struct{}),
		ctx:          ctx,
		cancelFn:     cancel,
	}

	// Start the stream
	if err := stream.start(); err != nil {
		cancel()
		return "", fmt.Errorf("failed to start stream: %w", err)
	}

	s.streams[cameraID] = stream

	log.Printf("Started camera stream for camera %d", cameraID)
	return s.GetCameraStreamURL(c, cameraID), nil
}

// StopCameraStream stops a stream for a specific camera
func (s *CameraStreamService) StopCameraStream(cameraID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream, exists := s.streams[cameraID]
	if !exists || !stream.isRunning {
		return fmt.Errorf("no active stream for camera %d", cameraID)
	}

	stream.stop()
	delete(s.streams, cameraID)

	log.Printf("Stopped camera stream for camera %d", cameraID)
	return nil
}

// handleStream serves an MJPEG stream for a specific camera
func (s *CameraStreamService) handleStream(c *fiber.Ctx) error {
	// Get camera ID from URL parameters
	cameraID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid camera ID")
	}

	s.mu.Lock()
	stream, exists := s.streams[uint(cameraID)]
	if !exists || !stream.isRunning {
		// If stream doesn't exist, try to start it
		s.mu.Unlock()

		// Try to get camera and check if it's active
		camera, err := s.cameraService.GetCameraByID(c.Context(), uint(cameraID))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Camera not found")
		}

		if camera.Status != "active" {
			return c.Status(fiber.StatusBadRequest).SendString("Camera is not active")
		}

		// Start stream with default config
		config := StreamConfig{
			FrameRate: 10,
			Quality:   75,
		}

		_, err = s.StartCameraStream(c, uint(cameraID), config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to start stream: " + err.Error())
		}

		s.mu.Lock()
		stream = s.streams[uint(cameraID)]
	}
	s.mu.Unlock()

	// Prepare for streaming
	c.Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Connection", "close")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Pragma", "no-cache")

	// Let's use Fiber's SendStream function instead
	return c.SendStream(stream.createStreamReader())
}

// createStreamReader creates a reader that streams MJPEG frames
func (s *CameraStream) createStreamReader() io.Reader {
	// Create pipe for streaming data
	pipeReader, pipeWriter := io.Pipe()

	// Create a channel for this client
	frameChan := make(chan []byte, 10)

	// Register client
	s.registerClient(frameChan)

	// Goroutine to write frames to the pipe
	go func() {
		defer pipeWriter.Close()
		defer s.unregisterClient(frameChan)

		// Send initial frame if available
		if s.frameCache != nil {
			writeFrame(pipeWriter, s.frameCache)
		}

		for {
			select {
			case frame, ok := <-frameChan:
				if !ok {
					// Channel closed, exit
					return
				}

				if err := writeFrame(pipeWriter, frame); err != nil {
					return
				}

			case <-s.stopChan:
				return

			case <-s.ctx.Done():
				return
			}
		}
	}()

	return pipeReader
}

// writeFrame writes a single MJPEG frame to the writer
func writeFrame(w io.Writer, frameData []byte) error {
	boundary := fmt.Sprintf("--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frameData))

	if _, err := w.Write([]byte(boundary)); err != nil {
		return err
	}

	if _, err := w.Write(frameData); err != nil {
		return err
	}

	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}

	return nil
}

// handleImage serves a static image for a specific camera
func (s *CameraStreamService) handleImage(c *fiber.Ctx) error {
	// Get camera ID from URL parameters
	cameraID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid camera ID")
	}

	s.mu.Lock()
	stream, exists := s.streams[uint(cameraID)]
	if !exists || !stream.isRunning {
		// If stream doesn't exist, try to access the image directly
		imagePath := filepath.Join(s.baseImageDirectory, fmt.Sprintf("image_%d.jpg", cameraID))
		s.mu.Unlock()

		// Check if image exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("Camera image not found")
		}

		return c.SendFile(imagePath)
	}
	s.mu.Unlock()

	// If stream exists, use cached frame if available
	if stream.frameCache != nil {
		c.Set("Content-Type", "image/jpeg")
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return c.Send(stream.frameCache)
	}

	// Fallback to reading the image file
	return c.SendFile(stream.imagePath)
}

// start initializes and starts the frame generator for a camera stream
func (s *CameraStream) start() error {
	if s.isRunning {
		return nil
	}

	go s.frameGenerator()

	s.isRunning = true
	return nil
}

// stop shuts down the frame generator and cleans up resources
func (s *CameraStream) stop() {
	if !s.isRunning {
		return
	}

	close(s.stopChan)

	s.clientsMu.Lock()
	for client := range s.clients {
		close(client)
	}
	s.clients = make(map[chan []byte]bool)
	s.lastActivity = make(map[chan []byte]time.Time)
	s.clientsMu.Unlock()

	s.cancelFn()
	s.isRunning = false
}

// sendMJPEGFrame writes a MJPEG frame to the response writer
func (s *CameraStream) sendMJPEGFrame(w io.Writer, frame []byte) error {
	_, err := fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))
	if err != nil {
		return err
	}

	_, err = w.Write(frame)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "\r\n")
	return err
}

// registerClient adds a client to the stream
func (s *CameraStream) registerClient(client chan []byte) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[client] = true
	s.lastActivity[client] = time.Now()

	log.Printf("New client connected to camera %d stream", s.cameraID)
}

// unregisterClient removes a client from the stream
func (s *CameraStream) unregisterClient(client chan []byte) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		delete(s.lastActivity, client)
		close(client)

		log.Printf("Client disconnected from camera %d stream", s.cameraID)
	}
}

// cleanupInactiveClients removes clients that haven't been active
func (s *CameraStream) cleanupInactiveClients() {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	now := time.Now()
	timeout := 2 * time.Minute

	for client, lastActive := range s.lastActivity {
		if now.Sub(lastActive) > timeout {
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				delete(s.lastActivity, client)
				close(client)

				log.Printf("Removed inactive client from camera %d stream due to timeout", s.cameraID)
			}
		}
	}
}

// broadcastFrame sends a frame to all connected clients
func (s *CameraStream) broadcastFrame(frameData []byte) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	s.frameCache = frameData

	now := time.Now()
	for client := range s.clients {
		select {
		case client <- frameData:
			s.lastActivity[client] = now
		default:
			// Channel is full, skip this frame for this client
		}
	}
}

// frameGenerator reads image files and broadcasts frames
func (s *CameraStream) frameGenerator() {
	interval := time.Second / time.Duration(s.frameRate)

	buf := new(bytes.Buffer)
	buf.Grow(1024 * 1024) // Pre-allocate 1MB buffer

	cleanupTicker := time.NewTicker(30 * time.Second)
	defer cleanupTicker.Stop()

	log.Printf("Starting frame generator for camera %d: %s, %d fps, quality %d%%",
		s.cameraID, s.imagePath, s.frameRate, s.quality)

	for {
		select {
		case <-s.stopChan:
			log.Printf("Frame generator stopped for camera %d", s.cameraID)
			return
		case <-s.ctx.Done():
			log.Printf("Frame generator context canceled for camera %d", s.cameraID)
			return
		case <-cleanupTicker.C:
			s.cleanupInactiveClients()
		default:
			if _, err := os.Stat(s.imagePath); os.IsNotExist(err) {
				time.Sleep(interval)
				continue
			}

			file, err := os.Open(s.imagePath)
			if err != nil {
				time.Sleep(interval)
				continue
			}

			img, _, err := image.Decode(file)
			file.Close()
			if err != nil {
				time.Sleep(interval)
				continue
			}

			buf.Reset()

			err = jpeg.Encode(buf, img, &jpeg.Options{Quality: s.quality})
			if err != nil {
				time.Sleep(interval)
				continue
			}

			s.broadcastFrame(buf.Bytes())

			time.Sleep(interval)
		}
	}
}

// AutoStartAllCameraStreams starts streams for all active cameras
func (s *CameraStreamService) AutoStartAllCameraStreams(c *fiber.Ctx, config StreamConfig) (map[uint]string, error) {
	// Get all active cameras
	cameras, err := s.cameraService.GetAllCameras(c.Context(), "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get active cameras: %w", err)
	}

	results := make(map[uint]string)

	// Start a stream for each camera
	for _, camera := range cameras {
		url, err := s.StartCameraStream(c, camera.ID, config)
		if err != nil {
			log.Printf("Failed to start stream for camera %d: %v", camera.ID, err)
			continue
		}

		results[camera.ID] = url
	}

	return results, nil
}

// GetStreamInfo returns information about a camera stream
func (s *CameraStreamService) GetStreamInfo(c *fiber.Ctx, cameraID uint) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream, exists := s.streams[cameraID]
	if !exists {
		return nil, fmt.Errorf("no stream for camera %d", cameraID)
	}

	stream.clientsMu.Lock()
	clientCount := len(stream.clients)
	stream.clientsMu.Unlock()

	baseURL := c.Protocol() + "://" + c.Hostname()
	streamURL := fmt.Sprintf("%s/api/cameras/%d/stream", baseURL, cameraID)
	imageURL := fmt.Sprintf("%s/api/cameras/%d/image", baseURL, cameraID)

	return map[string]interface{}{
		"camera_id":  stream.cameraID,
		"frame_rate": stream.frameRate,
		"quality":    stream.quality,
		"image_path": stream.imagePath,
		"is_running": stream.isRunning,
		"clients":    clientCount,
		"stream_url": streamURL,
		"image_url":  imageURL,
	}, nil
}
