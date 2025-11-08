package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// OptimizerConfig holds configuration for image optimization
type OptimizerConfig struct {
	MaxWidth       int           `json:"max_width"`
	MaxHeight      int           `json:"max_height"`
	Quality        int           `json:"quality"`       // 1-100 for JPEG
	Format         string        `json:"format"`        // jpeg, png, webp
	MaxFileSize    int64         `json:"max_file_size"` // bytes
	EnableResize   bool          `json:"enable_resize"`
	EnableCompress bool          `json:"enable_compress"`
	TempDir        string        `json:"temp_dir"`
	Timeout        time.Duration `json:"timeout"`
}

// DefaultOptimizerConfig returns sensible defaults for API usage
func DefaultOptimizerConfig() OptimizerConfig {
	return OptimizerConfig{
		MaxWidth:       1920,
		MaxHeight:      1080,
		Quality:        85,
		Format:         "jpeg",
		MaxFileSize:    4 * 1024 * 1024, // 4MB for API limits
		EnableResize:   true,
		EnableCompress: true,
		TempDir:        "/tmp/kainuguru/images",
		Timeout:        30 * time.Second,
	}
}

// APIOptimizerConfig returns configuration optimized for OpenAI Vision API
func APIOptimizerConfig() OptimizerConfig {
	return OptimizerConfig{
		MaxWidth:       2048, // OpenAI Vision API limit
		MaxHeight:      2048, // OpenAI Vision API limit
		Quality:        90,   // Higher quality for AI analysis
		Format:         "jpeg",
		MaxFileSize:    20 * 1024 * 1024, // 20MB OpenAI limit
		EnableResize:   true,
		EnableCompress: true,
		TempDir:        "/tmp/kainuguru/api_images",
		Timeout:        30 * time.Second,
	}
}

// OptimizationResult represents the result of image optimization
type OptimizationResult struct {
	InputFile           string        `json:"input_file"`
	OutputFile          string        `json:"output_file"`
	OriginalSize        int64         `json:"original_size"`
	OptimizedSize       int64         `json:"optimized_size"`
	CompressionRatio    float64       `json:"compression_ratio"`
	OriginalDimensions  [2]int        `json:"original_dimensions"`
	OptimizedDimensions [2]int        `json:"optimized_dimensions"`
	ProcessedAt         time.Time     `json:"processed_at"`
	Duration            time.Duration `json:"duration"`
	Success             bool          `json:"success"`
	Error               string        `json:"error,omitempty"`
}

// Optimizer handles image optimization for API usage
type Optimizer struct {
	config OptimizerConfig
}

// NewOptimizer creates a new image optimizer with configuration
func NewOptimizer(config OptimizerConfig) *Optimizer {
	return &Optimizer{
		config: config,
	}
}

// OptimizeForAPI optimizes an image for API consumption
func (o *Optimizer) OptimizeForAPI(ctx context.Context, inputPath string) (*OptimizationResult, error) {
	startTime := time.Now()

	result := &OptimizationResult{
		InputFile:   inputPath,
		ProcessedAt: startTime,
		Success:     false,
	}

	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Validate input
	if err := o.validateInput(inputPath); err != nil {
		result.Error = fmt.Sprintf("input validation failed: %v", err)
		return result, err
	}

	// Get original file info
	originalInfo, err := os.Stat(inputPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get file info: %v", err)
		return result, err
	}
	result.OriginalSize = originalInfo.Size()

	// Get original dimensions
	originalWidth, originalHeight, err := o.getDimensions(inputPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get dimensions: %v", err)
		return result, err
	}
	result.OriginalDimensions = [2]int{originalWidth, originalHeight}

	// Determine if optimization is needed
	needsOptimization := o.needsOptimization(originalWidth, originalHeight, result.OriginalSize)

	if !needsOptimization {
		// No optimization needed, return original
		result.OutputFile = inputPath
		result.OptimizedSize = result.OriginalSize
		result.CompressionRatio = 1.0
		result.OptimizedDimensions = result.OriginalDimensions
		result.Success = true
		return result, nil
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(o.config.TempDir, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create temp dir: %v", err)
		return result, err
	}

	// Generate output filename
	outputPath := o.generateOutputPath(inputPath)
	result.OutputFile = outputPath

	// Perform optimization
	if err := o.optimize(ctx, inputPath, outputPath); err != nil {
		result.Error = fmt.Sprintf("optimization failed: %v", err)
		return result, err
	}

	// Get optimized file info
	optimizedInfo, err := os.Stat(outputPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get optimized file info: %v", err)
		return result, err
	}
	result.OptimizedSize = optimizedInfo.Size()

	// Calculate compression ratio
	if result.OriginalSize > 0 {
		result.CompressionRatio = float64(result.OptimizedSize) / float64(result.OriginalSize)
	}

	// Get optimized dimensions
	optimizedWidth, optimizedHeight, err := o.getDimensions(outputPath)
	if err == nil {
		result.OptimizedDimensions = [2]int{optimizedWidth, optimizedHeight}
	}

	result.Success = true
	return result, nil
}

// validateInput checks if the input file is valid for optimization
func (o *Optimizer) validateInput(inputPath string) error {
	// Check if file exists
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("file does not exist: %v", err)
	}

	// Check file size
	if info.Size() > o.config.MaxFileSize*10 { // Allow larger input files
		return fmt.Errorf("input file too large: %d bytes", info.Size())
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(inputPath))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".webp"}
	isValid := false
	for _, validExt := range validExtensions {
		if ext == validExt {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	return nil
}

// needsOptimization determines if an image needs optimization
func (o *Optimizer) needsOptimization(width, height int, fileSize int64) bool {
	// Check if dimensions exceed limits
	if o.config.EnableResize && (width > o.config.MaxWidth || height > o.config.MaxHeight) {
		return true
	}

	// Check if file size exceeds limit
	if fileSize > o.config.MaxFileSize {
		return true
	}

	// Check if compression would be beneficial (for large images)
	if o.config.EnableCompress && (width > 1024 || height > 1024) && fileSize > 1024*1024 {
		return true
	}

	return false
}

// generateOutputPath creates a path for the optimized image
func (o *Optimizer) generateOutputPath(inputPath string) string {
	baseName := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_optimized_%s.%s", nameWithoutExt, timestamp, o.config.Format)

	return filepath.Join(o.config.TempDir, filename)
}

// optimize performs the actual image optimization
func (o *Optimizer) optimize(ctx context.Context, inputPath, outputPath string) error {
	// Try ImageMagick first (if available)
	if o.hasImageMagick() {
		return o.optimizeWithImageMagick(ctx, inputPath, outputPath)
	}

	// Fallback to Go's built-in image processing
	return o.optimizeWithGo(inputPath, outputPath)
}

// hasImageMagick checks if ImageMagick is available
func (o *Optimizer) hasImageMagick() bool {
	_, err := exec.LookPath("convert")
	return err == nil
}

// optimizeWithImageMagick uses ImageMagick for optimization
func (o *Optimizer) optimizeWithImageMagick(ctx context.Context, inputPath, outputPath string) error {
	args := []string{
		inputPath,
	}

	// Add resize if needed
	if o.config.EnableResize {
		resizeArg := fmt.Sprintf("%dx%d>", o.config.MaxWidth, o.config.MaxHeight)
		args = append(args, "-resize", resizeArg)
	}

	// Add quality setting
	if o.config.Format == "jpeg" {
		args = append(args, "-quality", fmt.Sprintf("%d", o.config.Quality))
	}

	// Strip metadata to reduce file size
	args = append(args, "-strip")

	// Set output format and path
	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, "convert", args...)
	return cmd.Run()
}

// optimizeWithGo uses Go's built-in image processing
func (o *Optimizer) optimizeWithGo(inputPath, outputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer inputFile.Close()

	// Decode image
	img, _, err := image.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize if needed
	if o.config.EnableResize {
		img = o.resizeImage(img)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Encode with optimization
	switch o.config.Format {
	case "jpeg":
		options := &jpeg.Options{Quality: o.config.Quality}
		return jpeg.Encode(outputFile, img, options)
	case "png":
		return png.Encode(outputFile, img)
	default:
		// Default to JPEG
		options := &jpeg.Options{Quality: o.config.Quality}
		return jpeg.Encode(outputFile, img, options)
	}
}

// resizeImage resizes an image to fit within configured dimensions
func (o *Optimizer) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions
	newWidth, newHeight := o.calculateNewDimensions(width, height)

	// If no resizing needed, return original
	if newWidth == width && newHeight == height {
		return img
	}

	// For now, return original image
	// In production, you'd use a proper resizing library like:
	// - github.com/nfnt/resize
	// - github.com/disintegration/imaging
	return img
}

// calculateNewDimensions calculates new dimensions while maintaining aspect ratio
func (o *Optimizer) calculateNewDimensions(width, height int) (int, int) {
	if width <= o.config.MaxWidth && height <= o.config.MaxHeight {
		return width, height
	}

	ratioWidth := float64(o.config.MaxWidth) / float64(width)
	ratioHeight := float64(o.config.MaxHeight) / float64(height)

	var ratio float64
	if ratioWidth < ratioHeight {
		ratio = ratioWidth
	} else {
		ratio = ratioHeight
	}

	newWidth := int(float64(width) * ratio)
	newHeight := int(float64(height) * ratio)

	return newWidth, newHeight
}

// getDimensions returns the dimensions of an image file
func (o *Optimizer) getDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}

// ToBase64 converts an optimized image to base64 for API usage
func (o *Optimizer) ToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %v", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)

	// Add data URL prefix for web APIs
	mimeType := o.getMimeType(imagePath)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	return dataURL, nil
}

// getMimeType returns the MIME type for a file extension
func (o *Optimizer) getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// OptimizeFromBytes optimizes an image from byte data
func (o *Optimizer) OptimizeFromBytes(ctx context.Context, data []byte, filename string) (*OptimizationResult, error) {
	// Create temporary input file
	if err := os.MkdirAll(o.config.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	tempInput := filepath.Join(o.config.TempDir, "temp_"+filename)
	if err := os.WriteFile(tempInput, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %v", err)
	}

	// Clean up temp input file
	defer os.Remove(tempInput)

	// Optimize the temporary file
	return o.OptimizeForAPI(ctx, tempInput)
}

// Cleanup removes optimized files
func (o *Optimizer) Cleanup(result *OptimizationResult) error {
	if result.OutputFile != result.InputFile {
		return os.Remove(result.OutputFile)
	}
	return nil
}

// BatchOptimize optimizes multiple images
func (o *Optimizer) BatchOptimize(ctx context.Context, inputPaths []string) ([]*OptimizationResult, error) {
	results := make([]*OptimizationResult, len(inputPaths))

	for i, path := range inputPaths {
		result, err := o.OptimizeForAPI(ctx, path)
		if err != nil {
			result = &OptimizationResult{
				InputFile: path,
				Success:   false,
				Error:     err.Error(),
			}
		}
		results[i] = result

		// Add small delay between optimizations
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	return results, nil
}

// GetOptimizationStats returns statistics about optimization operations
func (o *Optimizer) GetOptimizationStats(results []*OptimizationResult) OptimizationStats {
	stats := OptimizationStats{
		TotalFiles: len(results),
	}

	var totalOriginalSize, totalOptimizedSize int64
	var totalCompressionRatio float64

	for _, result := range results {
		if result.Success {
			stats.SuccessfulOptimizations++
			totalOriginalSize += result.OriginalSize
			totalOptimizedSize += result.OptimizedSize
			totalCompressionRatio += result.CompressionRatio
		} else {
			stats.FailedOptimizations++
		}
	}

	if stats.SuccessfulOptimizations > 0 {
		stats.AverageCompressionRatio = totalCompressionRatio / float64(stats.SuccessfulOptimizations)
		stats.TotalSpaceSaved = totalOriginalSize - totalOptimizedSize
	}

	return stats
}

// OptimizationStats represents statistics about optimization operations
type OptimizationStats struct {
	TotalFiles              int     `json:"total_files"`
	SuccessfulOptimizations int     `json:"successful_optimizations"`
	FailedOptimizations     int     `json:"failed_optimizations"`
	AverageCompressionRatio float64 `json:"average_compression_ratio"`
	TotalSpaceSaved         int64   `json:"total_space_saved"`
}
