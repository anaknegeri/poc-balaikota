package synchronizer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileHandler interface {
	ProcessFile(filePath string) error
}

type FolderConfig struct {
	Path        string
	FilePattern string
	Handler     FileHandler
}

// SyncManager handles file synchronization using a file watcher approach
type SyncManager struct {
	// Watcher and context for file system events
	watcher     *fsnotify.Watcher
	watchCtx    context.Context
	watchCancel context.CancelFunc

	// Configuration and state
	folderConfigs map[string]FolderConfig
	pendingFiles  chan string
	running       bool

	// Lock for managing state
	mu sync.Mutex
}

// NewSyncManager creates a new  synchronizer
func NewSyncManager() (*SyncManager, error) {
	// Initialize watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	watchCtx, watchCancel := context.WithCancel(context.Background())

	return &SyncManager{
		watcher:       watcher,
		watchCtx:      watchCtx,
		watchCancel:   watchCancel,
		folderConfigs: make(map[string]FolderConfig),
		pendingFiles:  make(chan string, 100), // Buffer for pending files
		running:       false,
	}, nil
}

// AddFolder adds a folder to be watched with a specific handler
func (s *SyncManager) AddFolder(path string, pattern string, handler FileHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Make sure the folder exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Creating folder: %s\n", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", path, err)
		}
	}

	// Add to our configs
	s.folderConfigs[path] = FolderConfig{
		Path:        path,
		FilePattern: pattern,
		Handler:     handler,
	}

	// If already running, start watching now
	if s.running {
		if err := s.watcher.Add(path); err != nil {
			return fmt.Errorf("failed to watch folder %s: %w", path, err)
		}
		fmt.Printf("Now watching folder: %s\n", path)
	}

	return nil
}

// Start begins watching for file changes
func (s *SyncManager) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil // Already running
	}
	s.running = true
	s.mu.Unlock()

	fmt.Println("Starting sync manager...")

	// Add all configured folders to the watcher
	for path := range s.folderConfigs {
		if err := s.watcher.Add(path); err != nil {
			fmt.Printf("Failed to watch folder %s: %v\n", path, err)
			continue
		}
		fmt.Printf("Watching folder: %s", path)
	}

	// Start the event handling goroutine
	go s.watchEvents()

	// Start processing workers
	go s.processPendingFiles()

	return nil
}

// Stop halts synchronization and watching
func (s *SyncManager) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	fmt.Println("Stopping sync manager...")

	// Cancel the context to stop event handling
	s.watchCancel()

	// Create a new context for future use
	s.watchCtx, s.watchCancel = context.WithCancel(context.Background())

	// Close the watcher
	if s.watcher != nil {
		s.watcher.Close()
	}

	fmt.Println("Sync manager stopped")
	return nil
}

// watchEvents handles file watching events
func (s *SyncManager) watchEvents() {
	for {
		select {
		case <-s.watchCtx.Done():
			return

		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			// Handle the event
			s.handleWatchEvent(event)

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// handleWatchEvent processes a single file watcher event
func (s *SyncManager) handleWatchEvent(event fsnotify.Event) {
	// Only interested in create and write operations
	if event.Op&fsnotify.Create == 0 && event.Op&fsnotify.Write == 0 {
		return
	}

	// Get file info
	info, err := os.Stat(event.Name)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Failed to get info for %s: %v\n", event.Name, err)
		}
		return
	}

	// Skip directories
	if info.IsDir() {
		return
	}

	// Check if this file belongs to one of our watched folders
	s.mu.Lock()
	var belongsToFolder bool
	for folderPath, config := range s.folderConfigs {
		if filepath.Dir(event.Name) == folderPath && matchesPattern(event.Name, config.FilePattern) {
			belongsToFolder = true
			break
		}
	}
	s.mu.Unlock()

	if !belongsToFolder {
		return
	}

	// Queue the file for processing
	fmt.Printf("New/modified file detected: %s\n", event.Name)
	select {
	case s.pendingFiles <- event.Name:
		// Successfully queued
	default:
		fmt.Printf("Pending files queue is full, skipping %s\n", event.Name)
	}
}

// processPendingFiles handles the queue of files to be processed
func (s *SyncManager) processPendingFiles() {
	for {
		select {
		case <-s.watchCtx.Done():
			return

		case filePath := <-s.pendingFiles:
			s.processFile(filePath)
		}
	}
}

// processFile processes a single file with its designated handler
func (s *SyncManager) processFile(filePath string) {
	// Find the correct handler for this file
	s.mu.Lock()
	var handler FileHandler
	for folderPath, config := range s.folderConfigs {
		if filepath.Dir(filePath) == folderPath && matchesPattern(filePath, config.FilePattern) {
			handler = config.Handler
			break
		}
	}
	s.mu.Unlock()

	if handler == nil {
		fmt.Printf("No handler found for file: %s\n", filePath)
		return
	}

	// Process the file with its handler
	if err := handler.ProcessFile(filePath); err != nil {
		fmt.Printf("Error processing file %s: %v\n", filePath, err)
	}
}

// matchesPattern checks if a file matches the given pattern
func matchesPattern(file, pattern string) bool {
	match, err := filepath.Match(pattern, filepath.Base(file))
	if err != nil {
		return false
	}
	return match
}

// JSONHandler is a basic handler for JSON files
type JSONHandler struct {
	name      string
	processor func(filePath string, data map[string]interface{}) error
}

// GetName returns the name of this handler
func (h *JSONHandler) GetName() string {
	return h.name
}

// ProcessFile processes a JSON file
func (h *JSONHandler) ProcessFile(filePath string) error {
	// Wait a bit to ensure the file is completely written
	time.Sleep(100 * time.Millisecond)

	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process with the custom processor function
	if h.processor != nil {
		if err := h.processor(filePath, data); err != nil {
			return fmt.Errorf("processor error: %w", err)
		}
	}

	return nil
}

// NewJSONHandler creates a new JSON file handler
func NewJSONHandler(name string, processor func(filePath string, data map[string]interface{}) error) *JSONHandler {
	return &JSONHandler{
		name:      name,
		processor: processor,
	}
}
