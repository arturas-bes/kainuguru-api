package pdf

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ProcessorConfig holds configuration for PDF processing
type ProcessorConfig struct {
	TempDir      string        `json:"temp_dir"`
	DPI          int           `json:"dpi"`
	Format       string        `json:"format"`  // jpeg, png, ppm
	Quality      int           `json:"quality"` // 1-100 for jpeg
	Timeout      time.Duration `json:"timeout"`
	Cleanup      bool          `json:"cleanup"`
	DeleteSource bool          `json:"delete_source"` // Delete source PDF after successful conversion
	MaxFileSize  int64         `json:"max_file_size"` // bytes
}

// DefaultProcessorConfig returns sensible defaults
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		TempDir:      "/tmp/kainuguru/pdf",
		DPI:          200,
		Format:       "jpeg",
		Quality:      85,
		Timeout:      30 * time.Second,
		Cleanup:      true,
		DeleteSource: false,            // Default to false for safety
		MaxFileSize:  50 * 1024 * 1024, // 50MB
	}
}

// ProcessingResult represents the result of PDF processing
type ProcessingResult struct {
	InputFile   string            `json:"input_file"`
	OutputFiles []string          `json:"output_files"`
	PageCount   int               `json:"page_count"`
	ProcessedAt time.Time         `json:"processed_at"`
	Duration    time.Duration     `json:"duration"`
	Success     bool              `json:"success"`
	Error       string            `json:"error,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Processor handles PDF to image conversion using pdftoppm
type Processor struct {
	config ProcessorConfig
}

// NewProcessor creates a new PDF processor with configuration
func NewProcessor(config ProcessorConfig) *Processor {
	return &Processor{
		config: config,
	}
}

// ProcessPDF converts a PDF file to images using pdftoppm
func (p *Processor) ProcessPDF(ctx context.Context, pdfPath string) (*ProcessingResult, error) {
	startTime := time.Now()

	result := &ProcessingResult{
		InputFile:   pdfPath,
		ProcessedAt: startTime,
		Success:     false,
	}

	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Validate input file
	if err := p.validateInputFile(pdfPath); err != nil {
		result.Error = fmt.Sprintf("input validation failed: %v", err)
		return result, err
	}

	// Ensure temp directory exists
	if err := p.ensureTempDir(); err != nil {
		result.Error = fmt.Sprintf("failed to create temp directory: %v", err)
		return result, err
	}

	// Get PDF metadata
	pageCount, metadata, err := p.getPDFMetadata(ctx, pdfPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get PDF metadata: %v", err)
		return result, err
	}

	result.PageCount = pageCount
	result.Metadata = metadata

	// Convert PDF to images
	outputFiles, err := p.convertToImages(ctx, pdfPath, pageCount)
	if err != nil {
		result.Error = fmt.Sprintf("PDF conversion failed: %v", err)
		return result, err
	}

	result.OutputFiles = outputFiles
	result.Success = true

	// Delete source PDF if configured and conversion was successful
	if p.config.DeleteSource {
		if err := os.Remove(pdfPath); err != nil {
			// Log warning but don't fail - conversion was successful
			fmt.Fprintf(os.Stderr, "Warning: failed to delete source PDF %s: %v\n", pdfPath, err)
		}
	}

	return result, nil
}

// validateInputFile checks if the input PDF file is valid
func (p *Processor) validateInputFile(pdfPath string) error {
	// Check if file exists
	info, err := os.Stat(pdfPath)
	if err != nil {
		return fmt.Errorf("file does not exist: %v", err)
	}

	// Check file size
	if info.Size() > p.config.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), p.config.MaxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(pdfPath))
	if ext != ".pdf" {
		return fmt.Errorf("invalid file extension: %s (expected .pdf)", ext)
	}

	return nil
}

// ensureTempDir creates the temporary directory if it doesn't exist
func (p *Processor) ensureTempDir() error {
	return os.MkdirAll(p.config.TempDir, 0755)
}

// getPDFMetadata extracts metadata from the PDF using pdfinfo
func (p *Processor) getPDFMetadata(ctx context.Context, pdfPath string) (int, map[string]string, error) {
	// Use pdfinfo to get PDF metadata
	cmd := exec.CommandContext(ctx, "pdfinfo", pdfPath)
	output, err := cmd.Output()
	if err != nil {
		// If pdfinfo is not available, try alternative method
		return p.getPageCountAlternative(ctx, pdfPath)
	}

	metadata := make(map[string]string)
	pageCount := 0

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				metadata[key] = value

				if key == "Pages" {
					if count, err := strconv.Atoi(value); err == nil {
						pageCount = count
					}
				}
			}
		}
	}

	if pageCount == 0 {
		return 0, metadata, fmt.Errorf("could not determine page count")
	}

	return pageCount, metadata, nil
}

// getPageCountAlternative tries to get page count without pdfinfo
func (p *Processor) getPageCountAlternative(ctx context.Context, pdfPath string) (int, map[string]string, error) {
	// Try using pdftoppm to get page count
	cmd := exec.CommandContext(ctx, "pdftoppm", "-l", "1", pdfPath, "/dev/null")
	if err := cmd.Run(); err != nil {
		// If even pdftoppm fails, assume 1 page
		return 1, make(map[string]string), nil
	}

	// For now, return a default value
	// In production, you might want to implement a more sophisticated method
	metadata := map[string]string{
		"Producer": "Unknown",
		"Creator":  "Unknown",
	}

	return 10, metadata, nil // Default assumption
}

// convertToImages converts PDF pages to images using pdftoppm
func (p *Processor) convertToImages(ctx context.Context, pdfPath string, pageCount int) ([]string, error) {
	// Check if pdftoppm is available
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm not found: %v", err)
	}

	// Generate output filename prefix
	baseName := filepath.Base(pdfPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputPrefix := filepath.Join(p.config.TempDir, nameWithoutExt)

	// Clean up any existing output files with the same prefix
	if err := p.cleanupOldFiles(outputPrefix); err != nil {
		// Log but don't fail - this is a best-effort cleanup
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup old files: %v\n", err)
	}

	// Build pdftoppm command
	args := []string{
		"-" + p.config.Format,
		"-r", strconv.Itoa(p.config.DPI),
	}

	// Add quality for JPEG
	if p.config.Format == "jpeg" && p.config.Quality > 0 {
		args = append(args, "-jpegopt", fmt.Sprintf("quality=%d", p.config.Quality))
	}

	// Add input and output
	args = append(args, pdfPath, outputPrefix)

	// Create command with timeout
	cmd := exec.CommandContext(ctx, "pdftoppm", args...)

	// Run the conversion
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm execution failed: %v", err)
	}

	// Find generated files
	outputFiles, err := p.findOutputFiles(outputPrefix, pageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to find output files: %v", err)
	}

	return outputFiles, nil
}

// findOutputFiles locates the generated image files
func (p *Processor) findOutputFiles(outputPrefix string, expectedCount int) ([]string, error) {
	var outputFiles []string

	// Get file extension based on format
	ext := p.config.Format
	if ext == "jpeg" {
		ext = "jpg" // pdftoppm uses .jpg for jpeg format
	}

	// pdftoppm generates files with different patterns based on page count
	// Try 2-digit numbering first: prefix-01.jpg, prefix-02.jpg, etc.
	for i := 1; i <= expectedCount; i++ {
		filename := fmt.Sprintf("%s-%02d.%s", outputPrefix, i, ext)
		if _, err := os.Stat(filename); err == nil {
			outputFiles = append(outputFiles, filename)
		}
	}

	// If 2-digit numbering didn't work, try 3-digit numbering
	if len(outputFiles) == 0 {
		for i := 1; i <= expectedCount; i++ {
			filename := fmt.Sprintf("%s-%03d.%s", outputPrefix, i, ext)
			if _, err := os.Stat(filename); err == nil {
				outputFiles = append(outputFiles, filename)
			}
		}
	}

	// If no numbered files found, check for single page output
	if len(outputFiles) == 0 {
		singleFile := fmt.Sprintf("%s.%s", outputPrefix, ext)
		if _, err := os.Stat(singleFile); err == nil {
			outputFiles = append(outputFiles, singleFile)
		}
	}

	if len(outputFiles) == 0 {
		return nil, fmt.Errorf("no output files generated")
	}

	return outputFiles, nil
}

// ProcessFromReader processes a PDF from an io.Reader
func (p *Processor) ProcessFromReader(ctx context.Context, reader io.Reader, filename string) (*ProcessingResult, error) {
	// Create temporary file
	tempFile := filepath.Join(p.config.TempDir, filename)
	if err := p.ensureTempDir(); err != nil {
		return nil, err
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	// Copy data to temp file
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to write temp file: %v", err)
	}

	// Process the temporary file
	result, err := p.ProcessPDF(ctx, tempFile)

	// Clean up temp file if configured (note: DeleteSource in ProcessPDF might have already deleted it)
	if p.config.Cleanup && !p.config.DeleteSource {
		os.Remove(tempFile)
	}

	return result, err
}

// cleanupOldFiles removes old output files that might conflict
func (p *Processor) cleanupOldFiles(outputPrefix string) error {
	dir := filepath.Dir(outputPrefix)
	base := filepath.Base(outputPrefix)

	// Get file extension
	ext := p.config.Format
	if ext == "jpeg" {
		ext = "jpg"
	}

	// Remove any existing files matching the pattern
	pattern := fmt.Sprintf("%s-*.%s", base, ext)
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return err
	}

	for _, match := range matches {
		os.Remove(match) // Best effort, ignore errors
	}

	// Also remove single file if exists
	singleFile := fmt.Sprintf("%s.%s", outputPrefix, ext)
	os.Remove(singleFile)

	return nil
}

// Cleanup removes generated image files
func (p *Processor) Cleanup(result *ProcessingResult) error {
	if !p.config.Cleanup {
		return nil
	}

	var errors []string
	for _, file := range result.OutputFiles {
		if err := os.Remove(file); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetImageDimensions returns the dimensions of generated images
func (p *Processor) GetImageDimensions(imagePath string) (int, int, error) {
	// Use imagemagick identify if available
	cmd := exec.Command("identify", "-format", "%w %h", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get image dimensions: %v", err)
	}

	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid dimension output")
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %v", err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %v", err)
	}

	return width, height, nil
}

// IsToolAvailable checks if required tools are available
func IsToolAvailable() bool {
	_, err := exec.LookPath("pdftoppm")
	return err == nil
}

// GetToolVersion returns the version of pdftoppm
func GetToolVersion() (string, error) {
	cmd := exec.Command("pdftoppm", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
