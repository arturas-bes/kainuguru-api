package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanerService handles cleanup of archived data and orphaned files
type CleanerService interface {
	// CleanOrphanedImages removes images that are no longer referenced
	CleanOrphanedImages(ctx context.Context, dryRun bool) (*CleanupResult, error)

	// CleanArchivedFlyerImages removes images from archived flyers
	CleanArchivedFlyerImages(ctx context.Context, archiveIDs []string, dryRun bool) (*CleanupResult, error)

	// CleanExpiredArchives removes archives past their retention period
	CleanExpiredArchives(ctx context.Context, dryRun bool) (*CleanupResult, error)

	// CleanTempFiles removes temporary files older than specified duration
	CleanTempFiles(ctx context.Context, olderThan time.Duration, dryRun bool) (*CleanupResult, error)

	// ValidateImageReferences checks for broken image references
	ValidateImageReferences(ctx context.Context) (*ValidationResult, error)

	// OptimizeStorage performs storage optimization operations
	OptimizeStorage(ctx context.Context, options *OptimizationOptions) (*OptimizationResult, error)

	// GetCleanupStatistics returns statistics about cleanup operations
	GetCleanupStatistics(ctx context.Context) (*CleanupStatistics, error)

	// ScheduleRegularCleanup sets up regular cleanup jobs
	ScheduleRegularCleanup(ctx context.Context, config *CleanupConfig) error
}

// CleanupResult contains results of a cleanup operation
type CleanupResult struct {
	OperationType   string                 `json:"operation_type"`
	FilesProcessed  int                    `json:"files_processed"`
	FilesDeleted    int                    `json:"files_deleted"`
	BytesFreed      int64                  `json:"bytes_freed"`
	Duration        time.Duration          `json:"duration"`
	DryRun          bool                   `json:"dry_run"`
	Errors          []string               `json:"errors"`
	Details         map[string]interface{} `json:"details"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ValidationResult contains results of image reference validation
type ValidationResult struct {
	TotalReferences    int            `json:"total_references"`
	ValidReferences    int            `json:"valid_references"`
	BrokenReferences   int            `json:"broken_references"`
	OrphanedFileCount  int            `json:"orphaned_files"`
	BrokenFiles        []BrokenFile   `json:"broken_files"`
	OrphanedFiles      []OrphanedFile `json:"orphaned_file_list"`
	RecommendedActions []string       `json:"recommended_actions"`
	ValidationTime     time.Time      `json:"validation_time"`
}

// BrokenFile represents a file reference that's broken
type BrokenFile struct {
	FilePath     string `json:"file_path"`
	ReferencedBy string `json:"referenced_by"`
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Error        string `json:"error"`
}

// OrphanedFile represents a file that's no longer referenced
type OrphanedFile struct {
	FilePath    string    `json:"file_path"`
	Size        int64     `json:"size"`
	ModifiedAt  time.Time `json:"modified_at"`
	FileType    string    `json:"file_type"`
	Suggestions []string  `json:"suggestions"`
}

// OptimizationOptions contains options for storage optimization
type OptimizationOptions struct {
	CompressImages    bool    `json:"compress_images"`
	ResizeImages      bool    `json:"resize_images"`
	MaxImageWidth     int     `json:"max_image_width"`
	MaxImageHeight    int     `json:"max_image_height"`
	ImageQuality      int     `json:"image_quality"`
	ConvertToWebP     bool    `json:"convert_to_webp"`
	RemoveDuplicates  bool    `json:"remove_duplicates"`
	DedupeThreshold   float64 `json:"dedupe_threshold"`
	DryRun            bool    `json:"dry_run"`
}

// OptimizationResult contains results of storage optimization
type OptimizationResult struct {
	ImagesProcessed    int                    `json:"images_processed"`
	ImagesCompressed   int                    `json:"images_compressed"`
	ImagesResized      int                    `json:"images_resized"`
	DuplicatesRemoved  int                    `json:"duplicates_removed"`
	SpaceSaved         int64                  `json:"space_saved_bytes"`
	Duration           time.Duration          `json:"duration"`
	Details            map[string]interface{} `json:"details"`
	Errors             []string               `json:"errors"`
	CreatedAt          time.Time              `json:"created_at"`
}

// CleanupStatistics contains statistics about cleanup operations
type CleanupStatistics struct {
	TotalCleanupOperations  int                      `json:"total_cleanup_operations"`
	TotalFilesDeleted       int                      `json:"total_files_deleted"`
	TotalBytesFreed         int64                    `json:"total_bytes_freed"`
	CleanupsByType          map[string]int           `json:"cleanups_by_type"`
	BytesFreedByType        map[string]int64         `json:"bytes_freed_by_type"`
	AverageCleanupDuration  time.Duration            `json:"average_cleanup_duration"`
	LastCleanupTime         *time.Time               `json:"last_cleanup_time"`
	RecentOperations        []*CleanupResult         `json:"recent_operations"`
	StorageUtilization      *StorageUtilization      `json:"storage_utilization"`
	ScheduledCleanups       []*ScheduledCleanup      `json:"scheduled_cleanups"`
}

// StorageUtilization contains storage utilization information
type StorageUtilization struct {
	TotalSpace      int64     `json:"total_space"`
	UsedSpace       int64     `json:"used_space"`
	FreeSpace       int64     `json:"free_space"`
	UsagePercent    float64   `json:"usage_percent"`
	ImageStorage    int64     `json:"image_storage"`
	ArchiveStorage  int64     `json:"archive_storage"`
	TempStorage     int64     `json:"temp_storage"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ScheduledCleanup represents a scheduled cleanup operation
type ScheduledCleanup struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Schedule    string        `json:"schedule"`
	Enabled     bool          `json:"enabled"`
	LastRun     *time.Time    `json:"last_run"`
	NextRun     time.Time     `json:"next_run"`
	Config      *CleanupConfig `json:"config"`
}

// CleanupConfig contains configuration for cleanup operations
type CleanupConfig struct {
	Type                 string        `json:"type"`
	Schedule             string        `json:"schedule"`
	Enabled              bool          `json:"enabled"`
	RetentionPeriod      time.Duration `json:"retention_period"`
	MaxFileAge           time.Duration `json:"max_file_age"`
	DryRun               bool          `json:"dry_run"`
	NotifyOnCompletion   bool          `json:"notify_on_completion"`
	NotifyOnError        bool          `json:"notify_on_error"`
	MaxFilesToProcess    int           `json:"max_files_to_process"`
	MaxBytesToProcess    int64         `json:"max_bytes_to_process"`
	SkipRecentlyAccessed bool          `json:"skip_recently_accessed"`
}

// FileStorage interface for file operations
type FileStorage interface {
	Exists(path string) bool
	Delete(path string) error
	GetSize(path string) (int64, error)
	GetModTime(path string) (time.Time, error)
	List(dir string) ([]string, error)
	GetStats(path string) (os.FileInfo, error)
}

// Repository interfaces for database operations
type ImageRepository interface {
	GetAllImageReferences(ctx context.Context) ([]*ImageReference, error)
	GetOrphanedImagePaths(ctx context.Context, existingPaths []string) ([]string, error)
	UpdateImagePath(ctx context.Context, oldPath, newPath string) error
	DeleteImageReference(ctx context.Context, path string) error
}

// ImageReference represents an image reference in the database
type ImageReference struct {
	ID         int64  `json:"id"`
	EntityType string `json:"entity_type"` // "flyer", "product", "store"
	EntityID   int    `json:"entity_id"`
	ImageURL   string `json:"image_url"`
	ImagePath  string `json:"image_path"`
}

// cleanerService implements CleanerService
type cleanerService struct {
	storage         FileStorage
	imageRepo       ImageRepository
	flyerRepo       FlyerRepository
	productRepo     ProductRepository
	config          *CleanerServiceConfig
}

// CleanerServiceConfig contains service configuration
type CleanerServiceConfig struct {
	ImageBasePath      string
	TempBasePath       string
	ArchiveBasePath    string
	MaxFileAge         time.Duration
	MaxTempFileAge     time.Duration
	ChunkSize          int
	ConcurrentWorkers  int
	ValidationEnabled  bool
}

// NewCleanerService creates a new cleaner service
func NewCleanerService(
	storage FileStorage,
	imageRepo ImageRepository,
	flyerRepo FlyerRepository,
	productRepo ProductRepository,
	config *CleanerServiceConfig,
) CleanerService {
	return &cleanerService{
		storage:     storage,
		imageRepo:   imageRepo,
		flyerRepo:   flyerRepo,
		productRepo: productRepo,
		config:      config,
	}
}

func (s *cleanerService) CleanOrphanedImages(ctx context.Context, dryRun bool) (*CleanupResult, error) {
	startTime := time.Now()
	result := &CleanupResult{
		OperationType: "orphaned_images",
		DryRun:        dryRun,
		Details:       make(map[string]interface{}),
		CreatedAt:     startTime,
	}

	// Get all image references from database
	references, err := s.imageRepo.GetAllImageReferences(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get image references: %w", err)
	}

	// Build set of referenced paths
	referencedPaths := make(map[string]bool)
	for _, ref := range references {
		if ref.ImagePath != "" {
			referencedPaths[ref.ImagePath] = true
		}
	}

	// Find all image files on disk
	imageFiles, err := s.getAllImageFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list image files: %w", err)
	}

	// Find orphaned files
	orphanedFiles := []string{}
	var totalSize int64

	for _, imagePath := range imageFiles {
		if !referencedPaths[imagePath] {
			orphanedFiles = append(orphanedFiles, imagePath)
			if size, err := s.storage.GetSize(imagePath); err == nil {
				totalSize += size
			}
		}
	}

	result.FilesProcessed = len(imageFiles)
	result.Details["orphaned_files"] = orphanedFiles
	result.Details["referenced_files"] = len(referencedPaths)

	if !dryRun {
		// Delete orphaned files
		deletedCount := 0
		var deletedSize int64

		for _, filePath := range orphanedFiles {
			if size, err := s.storage.GetSize(filePath); err == nil {
				deletedSize += size
			}

			if err := s.storage.Delete(filePath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete %s: %v", filePath, err))
			} else {
				deletedCount++
			}
		}

		result.FilesDeleted = deletedCount
		result.BytesFreed = deletedSize
	} else {
		result.FilesDeleted = len(orphanedFiles)
		result.BytesFreed = totalSize
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (s *cleanerService) CleanArchivedFlyerImages(ctx context.Context, archiveIDs []string, dryRun bool) (*CleanupResult, error) {
	startTime := time.Now()
	result := &CleanupResult{
		OperationType: "archived_flyer_images",
		DryRun:        dryRun,
		Details:       make(map[string]interface{}),
		CreatedAt:     startTime,
	}

	// This would implement logic to clean images from archived flyers
	// For now, return minimal implementation
	result.Duration = time.Since(startTime)
	result.Details["archive_ids"] = archiveIDs

	return result, nil
}

func (s *cleanerService) CleanExpiredArchives(ctx context.Context, dryRun bool) (*CleanupResult, error) {
	startTime := time.Now()
	result := &CleanupResult{
		OperationType: "expired_archives",
		DryRun:        dryRun,
		Details:       make(map[string]interface{}),
		CreatedAt:     startTime,
	}

	// List all archive files
	archiveFiles, err := s.storage.List(s.config.ArchiveBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list archive files: %w", err)
	}

	expiredFiles := []string{}
	var totalSize int64
	cutoffTime := time.Now().Add(-s.config.MaxFileAge)

	for _, archiveFile := range archiveFiles {
		modTime, err := s.storage.GetModTime(archiveFile)
		if err != nil {
			continue
		}

		if modTime.Before(cutoffTime) {
			expiredFiles = append(expiredFiles, archiveFile)
			if size, err := s.storage.GetSize(archiveFile); err == nil {
				totalSize += size
			}
		}
	}

	result.FilesProcessed = len(archiveFiles)
	result.Details["expired_files"] = expiredFiles

	if !dryRun {
		deletedCount := 0
		var deletedSize int64

		for _, filePath := range expiredFiles {
			if size, err := s.storage.GetSize(filePath); err == nil {
				deletedSize += size
			}

			if err := s.storage.Delete(filePath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete %s: %v", filePath, err))
			} else {
				deletedCount++
			}
		}

		result.FilesDeleted = deletedCount
		result.BytesFreed = deletedSize
	} else {
		result.FilesDeleted = len(expiredFiles)
		result.BytesFreed = totalSize
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (s *cleanerService) CleanTempFiles(ctx context.Context, olderThan time.Duration, dryRun bool) (*CleanupResult, error) {
	startTime := time.Now()
	result := &CleanupResult{
		OperationType: "temp_files",
		DryRun:        dryRun,
		Details:       make(map[string]interface{}),
		CreatedAt:     startTime,
	}

	tempFiles, err := s.storage.List(s.config.TempBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list temp files: %w", err)
	}

	oldFiles := []string{}
	var totalSize int64
	cutoffTime := time.Now().Add(-olderThan)

	for _, tempFile := range tempFiles {
		modTime, err := s.storage.GetModTime(tempFile)
		if err != nil {
			continue
		}

		if modTime.Before(cutoffTime) {
			oldFiles = append(oldFiles, tempFile)
			if size, err := s.storage.GetSize(tempFile); err == nil {
				totalSize += size
			}
		}
	}

	result.FilesProcessed = len(tempFiles)
	result.Details["old_files"] = oldFiles

	if !dryRun {
		deletedCount := 0
		var deletedSize int64

		for _, filePath := range oldFiles {
			if size, err := s.storage.GetSize(filePath); err == nil {
				deletedSize += size
			}

			if err := s.storage.Delete(filePath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete %s: %v", filePath, err))
			} else {
				deletedCount++
			}
		}

		result.FilesDeleted = deletedCount
		result.BytesFreed = deletedSize
	} else {
		result.FilesDeleted = len(oldFiles)
		result.BytesFreed = totalSize
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (s *cleanerService) ValidateImageReferences(ctx context.Context) (*ValidationResult, error) {
	references, err := s.imageRepo.GetAllImageReferences(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get image references: %w", err)
	}

	result := &ValidationResult{
		TotalReferences:  len(references),
		BrokenFiles:      []BrokenFile{},
		OrphanedFiles:    []OrphanedFile{},
		ValidationTime:   time.Now(),
	}

	validCount := 0
	brokenCount := 0

	for _, ref := range references {
		if ref.ImagePath == "" {
			continue
		}

		if s.storage.Exists(ref.ImagePath) {
			validCount++
		} else {
			brokenCount++
			brokenFile := BrokenFile{
				FilePath:     ref.ImagePath,
				ReferencedBy: ref.EntityType,
				EntityType:   ref.EntityType,
				EntityID:     fmt.Sprintf("%d", ref.EntityID),
				Error:        "File not found on disk",
			}
			result.BrokenFiles = append(result.BrokenFiles, brokenFile)
		}
	}

	result.ValidReferences = validCount
	result.BrokenReferences = brokenCount

	// Find orphaned files
	imageFiles, err := s.getAllImageFiles()
	if err == nil {
		referencedPaths := make(map[string]bool)
		for _, ref := range references {
			if ref.ImagePath != "" {
				referencedPaths[ref.ImagePath] = true
			}
		}

		for _, imagePath := range imageFiles {
			if !referencedPaths[imagePath] {
				orphanedFile := OrphanedFile{
					FilePath: imagePath,
					FileType: s.getFileType(imagePath),
				}

				if size, err := s.storage.GetSize(imagePath); err == nil {
					orphanedFile.Size = size
				}

				if modTime, err := s.storage.GetModTime(imagePath); err == nil {
					orphanedFile.ModifiedAt = modTime
				}

				orphanedFile.Suggestions = []string{"Delete if confirmed orphaned", "Check for missing database references"}
				result.OrphanedFiles = append(result.OrphanedFiles, orphanedFile)
			}
		}

		result.OrphanedFileCount = len(result.OrphanedFiles)
	}

	// Generate recommendations
	if result.BrokenReferences > 0 {
		result.RecommendedActions = append(result.RecommendedActions, "Update or remove broken image references")
	}
	if result.OrphanedFileCount > 0 {
		result.RecommendedActions = append(result.RecommendedActions, "Clean up orphaned image files")
	}

	return result, nil
}

func (s *cleanerService) OptimizeStorage(ctx context.Context, options *OptimizationOptions) (*OptimizationResult, error) {
	startTime := time.Now()
	result := &OptimizationResult{
		Details:   make(map[string]interface{}),
		CreatedAt: startTime,
	}

	// This would implement storage optimization logic
	// For now, return minimal implementation
	result.Duration = time.Since(startTime)
	result.Details["options"] = options

	return result, nil
}

func (s *cleanerService) GetCleanupStatistics(ctx context.Context) (*CleanupStatistics, error) {
	return &CleanupStatistics{
		TotalCleanupOperations: 0,
		TotalFilesDeleted:      0,
		TotalBytesFreed:        0,
		CleanupsByType:         make(map[string]int),
		BytesFreedByType:       make(map[string]int64),
		RecentOperations:       []*CleanupResult{},
		StorageUtilization:     &StorageUtilization{},
		ScheduledCleanups:      []*ScheduledCleanup{},
	}, nil
}

func (s *cleanerService) ScheduleRegularCleanup(ctx context.Context, config *CleanupConfig) error {
	// This would implement cleanup scheduling
	return fmt.Errorf("scheduling not yet implemented")
}

// Helper methods

func (s *cleanerService) getAllImageFiles() ([]string, error) {
	return s.storage.List(s.config.ImageBasePath)
}

func (s *cleanerService) getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "JPEG"
	case ".png":
		return "PNG"
	case ".webp":
		return "WebP"
	case ".gif":
		return "GIF"
	default:
		return "Unknown"
	}
}