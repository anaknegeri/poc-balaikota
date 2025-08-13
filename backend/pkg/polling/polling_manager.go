package polling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileHandler interface {
	ProcessFile(filePath string) error
	GetName() string
}

type FolderConfig struct {
	Path            string
	FilePattern     string
	Handler         FileHandler
	ProcessedFolder string
	FailedFolder    string
	AddTimestamp    bool
}

type PollingManager struct {
	folderConfigs    map[string]FolderConfig
	running          bool
	ctx              context.Context
	cancel           context.CancelFunc
	mu               sync.Mutex
	processedFiles   map[string]time.Time
	processingFiles  map[string]bool
	processedFilesMu sync.Mutex
	pollInterval     time.Duration
}

func NewPollingManager(pollInterval time.Duration) *PollingManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &PollingManager{
		folderConfigs:   make(map[string]FolderConfig),
		running:         false,
		ctx:             ctx,
		cancel:          cancel,
		processedFiles:  make(map[string]time.Time),
		processingFiles: make(map[string]bool),
		pollInterval:    pollInterval,
	}
}

func (m *PollingManager) AddFolder(path string, pattern string, handler FileHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Creating folder: %s\n", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", path, err)
		}
	}

	processedFolder := filepath.Join(path, "processed")
	failedFolder := filepath.Join(path, "failed")

	if err := os.MkdirAll(processedFolder, 0755); err != nil {
		return fmt.Errorf("failed to create processed folder %s: %w", processedFolder, err)
	}

	if err := os.MkdirAll(failedFolder, 0755); err != nil {
		return fmt.Errorf("failed to create failed folder %s: %w", failedFolder, err)
	}

	m.folderConfigs[path] = FolderConfig{
		Path:            path,
		FilePattern:     pattern,
		Handler:         handler,
		ProcessedFolder: processedFolder,
		FailedFolder:    failedFolder,
		AddTimestamp:    true,
	}

	fmt.Printf("Added folder: %s -> processed: %s, failed: %s\n", path, processedFolder, failedFolder)
	return nil
}

func (m *PollingManager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	fmt.Printf("Starting polling manager with interval %v\n", m.pollInterval)
	go m.pollFolders()
	go m.cleanupProcessedFiles()
	return nil
}

func (m *PollingManager) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = false
	m.mu.Unlock()

	fmt.Println("Stopping polling manager...")
	m.cancel()
	m.ctx, m.cancel = context.WithCancel(context.Background())
	fmt.Println("Polling manager stopped")
	return nil
}

func (m *PollingManager) pollFolders() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	m.checkFolders()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkFolders()
		}
	}
}

func (m *PollingManager) checkFolders() {
	m.mu.Lock()
	configs := make(map[string]FolderConfig)
	for k, v := range m.folderConfigs {
		configs[k] = v
	}
	m.mu.Unlock()

	for path, config := range configs {
		m.checkFolder(path, config)
	}
}

func (m *PollingManager) checkFolder(folderPath string, config FolderConfig) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", folderPath, err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		match, err := filepath.Match(config.FilePattern, file.Name())
		if err != nil || !match {
			continue
		}

		filePath := filepath.Join(folderPath, file.Name())

		m.processedFilesMu.Lock()
		_, processed := m.processedFiles[filePath]
		processing := m.processingFiles[filePath]
		m.processedFilesMu.Unlock()

		if processed || processing {
			continue
		}

		info, err := file.Info()
		if err != nil {
			fmt.Printf("Error getting file info for %s: %v\n", filePath, err)
			continue
		}

		if time.Since(info.ModTime()) < 2*time.Second {
			continue
		}

		m.processedFilesMu.Lock()
		m.processingFiles[filePath] = true
		m.processedFilesMu.Unlock()

		go func(path string, handler FileHandler, cfg FolderConfig) {
			err := handler.ProcessFile(path)

			m.processedFilesMu.Lock()
			m.processedFiles[path] = time.Now()
			delete(m.processingFiles, path)
			m.processedFilesMu.Unlock()

			if err != nil {
				fmt.Printf("Error processing file %s: %v\n", path, err)
				m.moveFile(path, cfg.FailedFolder)
			} else {
				m.moveFile(path, cfg.ProcessedFolder)
			}
		}(filePath, config.Handler, config)
	}
}

func (m *PollingManager) moveFile(srcPath, destFolder string) {
	if destFolder == "" {
		return
	}

	fileName := filepath.Base(srcPath)

	destPath := filepath.Join(destFolder, fileName)

	if err := m.copyFile(srcPath, destPath); err != nil {
		fmt.Printf("Error copying file %s to %s: %v\n", srcPath, destPath, err)
		return
	}

	if err := os.Remove(srcPath); err != nil {
		fmt.Printf("Error removing original file %s: %v\n", srcPath, err)
		return
	}

	// fmt.Printf("Moved file %s to %s\n", srcPath, destPath)
}

func (m *PollingManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (m *PollingManager) cleanupProcessedFiles() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.processedFilesMu.Lock()
			now := time.Now()
			for file, processTime := range m.processedFiles {
				if now.Sub(processTime) > 2*time.Hour {
					delete(m.processedFiles, file)
				}
			}
			m.processedFilesMu.Unlock()
		}
	}
}

type JSONHandler struct {
	name      string
	processor func(filePath string, data map[string]interface{}) error
}

func (h *JSONHandler) GetName() string {
	return h.name
}

func (h *JSONHandler) ProcessFile(filePath string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if h.processor != nil {
		if err := h.processor(filePath, data); err != nil {
			return fmt.Errorf("processor error: %w", err)
		}
	}

	return nil
}

func NewJSONHandler(name string, processor func(filePath string, data map[string]interface{}) error) *JSONHandler {
	return &JSONHandler{
		name:      name,
		processor: processor,
	}
}
