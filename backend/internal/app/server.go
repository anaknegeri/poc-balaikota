package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"people-counting/config"
	"people-counting/internal/handler"
	"people-counting/internal/repository/postgres"
	"people-counting/internal/service"
	"people-counting/pkg/database"
	"people-counting/pkg/polling"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
)

// alertTypeFolders defines the mapping between alert types and their folder paths
var alertTypeFolders = map[string]string{
	"restricted":                    "alert",
	"fall-detection":                "fall_log",
	"loitering":                     "loitering",
	"personal-protective-equipment": "safety_log",
}

// Server represents the API server
type Server struct {
	config           *config.Config
	db               *gorm.DB
	app              *fiber.App
	syncManager      *polling.PollingManager
	streamService    *service.CameraStreamService // Menambahkan field untuk menyimpan reference ke streamService
	webSocketService *service.WebSocketService    // WebSocket service untuk mengelola koneksi WebSocket
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

// Initialize sets up the server
func (s *Server) Initialize() error {
	// Connect to database
	dbConfig := &database.Config{
		DSN:             s.config.Database.DSN,
		MaxOpenConns:    s.config.Database.MaxOpenConns,
		MaxIdleConns:    s.config.Database.MaxIdleConns,
		ConnMaxLifetime: s.config.Database.ConnMaxLifetime,
	}

	var err error
	s.db, err = database.NewConnection(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Check if TimescaleDB is installed
	if !database.IsTimescaleDBInstalled(s.db) {
		log.Println("WARNING: TimescaleDB extension is not installed. Time-series functionality may not work correctly.")
	}

	// Run database migrations
	if err := database.MigrateDatabase(s.db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize Fiber app
	s.app = fiber.New(fiber.Config{
		AppName:      "People Counting API",
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		ErrorHandler: customErrorHandler,
	})

	// Register middleware
	s.registerMiddleware()

	// Register routes
	s.registerRoutes()

	// Initialize and configure Sync Manager
	if err := s.initializeSyncManager(); err != nil {
		return fmt.Errorf("failed to initialize sync manager: %w", err)
	}

	return nil
}

// initializeSyncManager sets up the file sync manager
func (s *Server) initializeSyncManager() error {
	// Create new sync manager
	s.syncManager = polling.NewPollingManager(5 * time.Second)

	// Ensure data directories exist and are absolute paths
	dataRootDir := s.config.DataDirectories.Root

	// If data root is not absolute, make it relative to the current directory
	if !filepath.IsAbs(dataRootDir) {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
		dataRootDir = filepath.Join(currentDir, dataRootDir)
	}

	// Create directories for various data types
	countingDataDir := filepath.Join(dataRootDir, s.config.DataDirectories.PeopleCountDir)
	faceRecognitionDataDir := filepath.Join(dataRootDir, s.config.DataDirectories.FaceRecognitionDir)
	vehicleCountDataDir := filepath.Join(dataRootDir, s.config.DataDirectories.VehicleCountDir)

	// Create alert directories for each alert type
	for _, folderPath := range alertTypeFolders {
		alertDir := filepath.Join(dataRootDir, folderPath)

		// Create alert type directory
		if err := os.MkdirAll(alertDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", alertDir, err)
		}

		// Create both "image" and "images" directories to support both naming conventions
		imageFolders := []string{"image", "images"}
		for _, imageFolder := range imageFolders {
			imagesDir := filepath.Join(alertDir, imageFolder)
			if err := os.MkdirAll(imagesDir, 0755); err != nil {
				return fmt.Errorf("failed to create images directory %s: %v", imagesDir, err)
			}
		}
	}

	// Create directories for other data types
	for _, dir := range []string{countingDataDir, faceRecognitionDataDir, vehicleCountDataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Get repositories - use the ones already created in registerRoutes to avoid duplication
	cameraRepository := postgres.NewCameraRepository(s.db)
	alertTypeRepository := postgres.NewAlertTypeRepository(s.db)
	alertRepository := postgres.NewAlertRepository(s.db)
	peopleCountRepository := postgres.NewPeopleCountRepository(s.db)
	faceRecognitionRepository := postgres.NewFaceRecognitionRepository(s.db)
	vehicleRepository := postgres.NewVehicleCountRepository(s.db)

	// Create services using the same repositories
	cameraService := service.NewCameraService(cameraRepository, dataRootDir)
	alertTypeService := service.NewAlertTypeService(alertTypeRepository)
	alertService := service.NewAlertService(alertRepository, alertTypeRepository, cameraRepository)
	peopleCountService := service.NewPeopleCountService(peopleCountRepository)
	faceRecognitionService := service.NewFaceRecognitionService(faceRecognitionRepository, cameraRepository)
	vehicleService := service.NewVehicleCountService(vehicleRepository)

	// Create handlers for file sync
	syncCountingHandler := handler.NewPeopleCountHandler(peopleCountService, cameraService)
	syncRecognitionHandler := handler.NewFaceRecognitionHandler(faceRecognitionService, cameraService)
	syncVehicleCountingHandler := handler.NewVehicleCountHandler(vehicleService, cameraService)

	// Add alert folders to watch - one handler per alert type folder
	for alertType, folderPath := range alertTypeFolders {
		alertDir := filepath.Join(dataRootDir, folderPath)
		syncAlertHandler := handler.NewAlertHandlerWithType(alertTypeService, alertService, cameraService, s.webSocketService, alertDir, alertType)

		if err := s.syncManager.AddFolder(alertDir, "*.json", syncAlertHandler); err != nil {
			return fmt.Errorf("failed to add alert folder %s for type %s: %v", alertDir, alertType, err)
		}
		log.Printf("Added alert folder watcher: %s -> %s", alertType, alertDir)
	}

	// Add other folders to watch
	if err := s.syncManager.AddFolder(countingDataDir, "*.json", syncCountingHandler); err != nil {
		return fmt.Errorf("failed to add people counting folder: %v", err)
	}

	if err := s.syncManager.AddFolder(faceRecognitionDataDir, "*.json", syncRecognitionHandler); err != nil {
		return fmt.Errorf("failed to add face recognition folder: %v", err)
	}

	if err := s.syncManager.AddFolder(vehicleCountDataDir, "*.json", syncVehicleCountingHandler); err != nil {
		return fmt.Errorf("failed to add vehicle count folder: %v", err)
	}

	// Start the sync manager
	if err := s.syncManager.Start(); err != nil {
		return fmt.Errorf("failed to start sync manager: %v", err)
	}

	log.Println("Sync manager started successfully")
	return nil
}

// registerMiddleware registers middleware for the server
func (s *Server) registerMiddleware() {
	s.app.Use(logger.New())
	s.app.Use(recover.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))
}

// registerRoutes registers all the API routes
func (s *Server) registerRoutes() {
	api := s.app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Initialize WebSocket handler
	webSocketHandler := handler.NewWebSocketHandler()

	// Initialize WebSocket service
	s.webSocketService = service.NewWebSocketService(webSocketHandler)

	// Set up repositories
	cameraRepository := postgres.NewCameraRepository(s.db)
	peopleCountRepository := postgres.NewPeopleCountRepository(s.db)
	alertTypeRepository := postgres.NewAlertTypeRepository(s.db)
	alertRepository := postgres.NewAlertRepository(s.db)
	faceRecognitionRepository := postgres.NewFaceRecognitionRepository(s.db)
	vehicleRepository := postgres.NewVehicleCountRepository(s.db)

	// Ensure stream directory exists
	streamDir := filepath.Join(s.config.DataDirectories.Root, s.config.DataDirectories.StreamDir)
	if err := os.MkdirAll(streamDir, 0755); err != nil {
		log.Printf("WARNING: Failed to create stream directory: %v", err)
	}

	// Set up services
	cameraService := service.NewCameraService(cameraRepository, streamDir)
	peopleCountService := service.NewPeopleCountService(peopleCountRepository)
	alertTypeService := service.NewAlertTypeService(alertTypeRepository)
	alertService := service.NewAlertService(alertRepository, alertTypeRepository, cameraRepository)
	faceRecognitionService := service.NewFaceRecognitionService(faceRecognitionRepository, cameraRepository)
	vehicleService := service.NewVehicleCountService(vehicleRepository)

	// Initialize camera stream service
	s.streamService = service.NewCameraStreamService(cameraService, streamDir)

	// Start the streaming service
	if err := s.streamService.Start(); err != nil {
		log.Printf("WARNING: Failed to start camera streaming service: %v", err)
	} else {
		// Register streaming routes
		s.streamService.RegisterRoutes(api)
		log.Println("Camera streaming service started successfully")
	}

	// Auto-start streams for active cameras
	go func() {
		time.Sleep(2 * time.Second) // Short delay to allow server startup
		log.Printf("Camera streaming service initialized and ready for connections")
	}()

	// Set up static file serving
	faceImagesPath := filepath.Join(s.config.DataDirectories.Root, s.config.DataDirectories.FaceRecognitionDir, "images")
	api.Static("/images/faces", faceImagesPath)

	// Create static routes for each alert type
	for alertType, folderPath := range alertTypeFolders {
		basePath := filepath.Join(s.config.DataDirectories.Root, folderPath)
		routePath := fmt.Sprintf("/images/alerts/%s", alertType)

		// Try both "image" and "images" folders, prioritize the one with files
		imageFolders := []string{"images", "image"} // Try "images" first
		var alertImagesPath string

		for _, folder := range imageFolders {
			testPath := filepath.Join(basePath, folder)
			if _, err := os.Stat(testPath); err == nil {
				// Check if folder has any files
				entries, err := os.ReadDir(testPath)
				if err == nil && len(entries) > 0 {
					// Folder exists and has files, use this one
					alertImagesPath = testPath
					break
				} else if alertImagesPath == "" {
					// Folder exists but might be empty, keep as fallback
					alertImagesPath = testPath
				}
			}
		}

		// If neither exists, default to "images" folder (plural)
		if alertImagesPath == "" {
			alertImagesPath = filepath.Join(basePath, "images")
		}

		api.Static(routePath, alertImagesPath)
		log.Printf("Serving alert images for %s from: %s", alertType, alertImagesPath)
	}

	// Keep backward compatibility - serve general alerts folder
	alertBasePath := filepath.Join(s.config.DataDirectories.Root, s.config.DataDirectories.AlertDir)
	var generalAlertImagesPath string

	// Try both "image" and "images" folders for general alerts, prioritize the one with files
	generalImageFolders := []string{"images", "image"} // Try "images" first
	for _, folder := range generalImageFolders {
		testPath := filepath.Join(alertBasePath, folder)
		if _, err := os.Stat(testPath); err == nil {
			// Check if folder has any files
			entries, err := os.ReadDir(testPath)
			if err == nil && len(entries) > 0 {
				// Folder exists and has files, use this one
				generalAlertImagesPath = testPath
				break
			} else if generalAlertImagesPath == "" {
				// Folder exists but might be empty, keep as fallback
				generalAlertImagesPath = testPath
			}
		}
	}

	// If neither exists, default to "images" folder (plural)
	if generalAlertImagesPath == "" {
		generalAlertImagesPath = filepath.Join(alertBasePath, "images")
	}

	api.Static("/images/alerts", generalAlertImagesPath)

	// Serve websocket test file
	api.Static("/", "./websocket_test.html")

	// Set up handlers
	cameraHandler := handler.NewCameraHandler(cameraService)
	peopleCountHandler := handler.NewPeopleCountHandler(peopleCountService, cameraService)
	alertTypeHandler := handler.NewAlertTypeHandler(alertTypeService)
	alertHandler := handler.NewAlertHandler(alertTypeService, alertService, cameraService, s.webSocketService)
	faceRecognitionHandler := handler.NewFaceRecognitionHandler(faceRecognitionService, cameraService)
	vehicleCountingHandler := handler.NewVehicleCountHandler(vehicleService, cameraService)

	// Register handler routes
	cameraHandler.RegisterRoutes(api)
	peopleCountHandler.RegisterRoutes(api)
	alertTypeHandler.RegisterRoutes(api)
	alertHandler.RegisterRoutes(api)
	faceRecognitionHandler.RegisterRoutes(api)
	vehicleCountingHandler.RegisterRoutes(api)
	webSocketHandler.RegisterRoutes(api)
}

// Run starts the server
func (s *Server) Run() error {
	// Channel to listen for errors coming from the listener
	serverShutdown := make(chan struct{})
	// Start server in a goroutine
	go func() {
		if err := s.app.Listen(":" + s.config.Server.Port); err != nil {
			log.Printf("Server shutdown error: %v", err)
			close(serverShutdown)
		}
	}()

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case <-quit:
		log.Println("Shutting down server...")
	case <-serverShutdown:
		log.Println("Server stopped unexpectedly")
	}

	// Stop the sync manager
	if s.syncManager != nil {
		if err := s.syncManager.Stop(); err != nil {
			log.Printf("Error stopping sync manager: %v", err)
		} else {
			log.Println("Sync manager stopped successfully")
		}
	}

	// Stop the stream service
	if s.streamService != nil {
		if err := s.streamService.Stop(); err != nil {
			log.Printf("Error stopping stream service: %v", err)
		} else {
			log.Println("Stream service stopped successfully")
		}
	}

	// WebSocket service will be stopped automatically when server shuts down
	if s.webSocketService != nil {
		log.Println("WebSocket service will be stopped with server shutdown")
	}

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server shutdown complete")
	return nil
}

// customErrorHandler is the custom error handler for Fiber
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default 500 status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Send JSON response
	return c.Status(code).JSON(fiber.Map{
		"error": true,
		"msg":   err.Error(),
	})
}
