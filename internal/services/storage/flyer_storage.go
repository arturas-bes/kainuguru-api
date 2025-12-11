package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/rs/zerolog/log"
)

// FlyerStorageService handles flyer image storage operations
type FlyerStorageService interface {
	SaveFlyerPage(ctx context.Context, flyer *models.Flyer, pageNumber int, imageData io.Reader) (string, error)
	GetFlyerPageURL(flyer *models.Flyer, pageNumber int, baseURL string) string
	GetFlyerPagePath(flyer *models.Flyer, pageNumber int) string
	DeleteFlyer(ctx context.Context, flyer *models.Flyer) error
	EnforceStorageLimit(ctx context.Context, storeCode string) error
	EnsureDirectoryExists(path string) error
}

type fileSystemStorage struct {
	basePath  string // e.g., "/path/to/kainuguru-public"
	publicURL string // e.g., "https://yourdomain.com"
}

// NewFileSystemStorage creates a new filesystem-based storage service
func NewFileSystemStorage(basePath, publicURL string) FlyerStorageService {
	return &fileSystemStorage{
		basePath:  basePath,
		publicURL: publicURL,
	}
}

// SaveFlyerPage saves a flyer page image to disk and returns the relative path
func (s *fileSystemStorage) SaveFlyerPage(ctx context.Context, flyer *models.Flyer, pageNumber int, imageData io.Reader) (string, error) {
	// Construct file path
	relativePath := flyer.GetImageBasePath()
	fullPath := filepath.Join(s.basePath, relativePath)

	// Create directories if they don't exist
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create directory")
	}

	// Create file
	fileName := fmt.Sprintf("page-%d.jpg", pageNumber)
	filePath := filepath.Join(fullPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create file")
	}
	defer file.Close()

	// Copy image data
	written, err := io.Copy(file, imageData)
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to write image data")
	}

	log.Info().
		Str("path", filePath).
		Int64("bytes", written).
		Int("page", pageNumber).
		Msg("Saved flyer page")

	// Return the full CDN URL if publicURL is configured, otherwise relative path
	relPath := fmt.Sprintf("%s/%s", relativePath, fileName)
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.publicURL, "/"), relPath), nil
	}
	return relPath, nil
}

// GetFlyerPageURL returns the public URL for a flyer page
func (s *fileSystemStorage) GetFlyerPageURL(flyer *models.Flyer, pageNumber int, baseURL string) string {
	relativePath := flyer.GetImageBasePath()
	fileName := fmt.Sprintf("page-%d.jpg", pageNumber)
	// Use provided baseURL or fall back to configured publicURL
	if baseURL == "" {
		baseURL = s.publicURL
	}
	return fmt.Sprintf("%s/%s/%s", baseURL, relativePath, fileName)
}

// GetFlyerPagePath returns the local filesystem path for a flyer page
func (s *fileSystemStorage) GetFlyerPagePath(flyer *models.Flyer, pageNumber int) string {
	relativePath := flyer.GetImageBasePath()
	fileName := fmt.Sprintf("page-%d.jpg", pageNumber)
	return filepath.Join(s.basePath, relativePath, fileName)
}

// DeleteFlyer removes all images for a flyer
func (s *fileSystemStorage) DeleteFlyer(ctx context.Context, flyer *models.Flyer) error {
	flyerPath := filepath.Join(s.basePath, flyer.GetImageBasePath())
	if err := os.RemoveAll(flyerPath); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to delete flyer directory")
	}

	log.Info().
		Str("path", flyerPath).
		Int("flyerId", flyer.ID).
		Msg("Deleted flyer directory")

	return nil
}

// EnforceStorageLimit keeps only the 2 newest flyers per store
func (s *fileSystemStorage) EnforceStorageLimit(ctx context.Context, storeCode string) error {
	storePath := filepath.Join(s.basePath, "flyers", strings.ToLower(storeCode))

	// Read all flyer folders for this store
	flyerFolders, err := os.ReadDir(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Store directory doesn't exist yet, that's fine
			return nil
		}
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to read store directory")
	}

	// Filter only directories and sort by date (folder name starts with date)
	var folders []string
	for _, folder := range flyerFolders {
		if folder.IsDir() {
			folders = append(folders, folder.Name())
		}
	}

	// Sort folders (date prefix ensures chronological order)
	sort.Strings(folders)

	// Keep only the 2 newest folders
	const maxFolders = 2
	if len(folders) > maxFolders {
		// Delete oldest folders
		for i := 0; i < len(folders)-maxFolders; i++ {
			oldFolderPath := filepath.Join(storePath, folders[i])
			if err := os.RemoveAll(oldFolderPath); err != nil {
				log.Error().
					Err(err).
					Str("path", oldFolderPath).
					Msg("Failed to delete old flyer")
			} else {
				log.Info().
					Str("path", oldFolderPath).
					Str("store", storeCode).
					Msg("Deleted old flyer (keeping only 2 newest)")
			}
		}
	}

	return nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func (s *fileSystemStorage) EnsureDirectoryExists(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create directory")
	}
	return nil
}
