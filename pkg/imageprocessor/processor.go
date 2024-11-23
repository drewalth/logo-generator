package imageprocessor

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

// Dimension represents the size and filename for the output image.
type Dimension struct {
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Name   string `json:"name"`
}

// ImageError provides a custom error type with more context.
type ImageError struct {
	Func string
	Msg  string
	Err  error
}

func (e *ImageError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Func, e.Msg, e.Err)
}

func wrapError(funcName, msg string, err error) error {
	return &ImageError{Func: funcName, Msg: msg, Err: err}
}

// LoadDimensions reads the dimensions configuration from a JSON file.
func LoadDimensions(configPath string) ([]Dimension, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, wrapError("LoadDimensions", "failed to read config file", err)
	}
	var dims []Dimension
	if err := json.Unmarshal(data, &dims); err != nil {
		return nil, wrapError("LoadDimensions", "failed to parse config file", err)
	}
	return dims, nil
}

// ProcessImage orchestrates the image processing with concurrency and caching.
func ProcessImage(ctx context.Context, inputPath, outputDir string, dimensions []Dimension, infoLogger, errorLogger *log.Logger) error {
	// Check if dimensions are provided
	if len(dimensions) == 0 {
		return wrapError("ProcessImage", "no dimensions specified for resizing", nil)
	}

	// Open the input image file
	file, err := os.Open(inputPath)
	if err != nil {
		return wrapError("ProcessImage", "failed to open image file", err)
	}
	defer file.Close()

	// Decode the image
	srcImg, format, err := image.Decode(file)
	if err != nil {
		return wrapError("ProcessImage", "failed to decode image", err)
	}

	// Validate supported formats
	if !isSupportedFormat(format) {
		return wrapError("ProcessImage", "unsupported image format", fmt.Errorf(format))
	}

	// Validate image dimensions
	bounds := srcImg.Bounds()
	if bounds.Dx() != 1080 || bounds.Dy() != 1080 {
		return wrapError("ProcessImage", "image dimensions must be 1080x1080", nil)
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return wrapError("ProcessImage", "failed to create output directory", err)
	}

	// Initialize cache directory
	cachePath := getCachePath(inputPath)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return wrapError("ProcessImage", "failed to create cache directory", err)
	}

	// Implement concurrency
	var wg sync.WaitGroup
	errChan := make(chan error, len(dimensions))

	for _, dim := range dimensions {
		select {
		case <-ctx.Done():
			return wrapError("ProcessImage", "processing canceled or timed out", ctx.Err())
		default:
			wg.Add(1)
			go func(dim Dimension) {
				defer wg.Done()
				outputPath := filepath.Join(outputDir, dim.Name)

				// Check cache
				if isCached(cachePath, dim.Name) {
					infoLogger.Printf("Cached: %s\n", dim.Name)
					return
				}

				// Resize and save image
				if err := resizeAndSaveImage(srcImg, dim.Width, dim.Height, outputPath); err != nil {
					errChan <- wrapError("resizeAndSaveImage", dim.Name, err)
					return
				}

				// Update cache
				if err := updateCache(cachePath, dim.Name); err != nil {
					errorLogger.Printf("Failed to update cache for %s: %v\n", dim.Name, err)
				}

				infoLogger.Printf("Processed: %s\n", dim.Name)
			}(dim)
		}
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	for err := range errChan {
		return err
	}

	return nil
}

// isSupportedFormat checks if the image format is supported.
func isSupportedFormat(format string) bool {
	switch strings.ToLower(format) {
	case "png", "jpeg", "gif":
		return true
	default:
		return false
	}
}

// resizeAndSaveImage resizes the image while maintaining aspect ratio and saves it.
func resizeAndSaveImage(src image.Image, width, height uint, outputPath string) error {
	// Calculate aspect ratio
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	var newWidth, newHeight uint
	if srcWidth > srcHeight {
		newWidth = width
		newHeight = uint(float64(srcHeight) * (float64(width) / float64(srcWidth)))
	} else {
		newHeight = height
		newWidth = uint(float64(srcWidth) * (float64(height) / float64(srcHeight)))
	}

	// Resize the image
	resizedImg := resize.Resize(newWidth, newHeight, src, resize.Lanczos3)

	// Create a new RGBA image with desired dimensions
	paddedImg := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	// Fill the background with transparent color
	draw.Draw(paddedImg, paddedImg.Bounds(), &image.Uniform{C: image.Transparent}, image.Point{}, draw.Src)

	// Calculate offset to center the image
	offsetX := (int(width) - resizedImg.Bounds().Dx()) / 2
	offsetY := (int(height) - resizedImg.Bounds().Dy()) / 2

	// Draw the resized image onto the padded image
	draw.Draw(paddedImg, resizedImg.Bounds().Add(image.Pt(offsetX, offsetY)), resizedImg, image.Point{}, draw.Over)

	// Determine the encoding format based on file extension
	extension := strings.ToLower(filepath.Ext(outputPath))
	var encodeFunc func(io.Writer, image.Image) error

	switch extension {
	case ".png":
		encodeFunc = png.Encode
	case ".jpg", ".jpeg":
		encodeFunc = func(w io.Writer, img image.Image) error {
			return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
		}
	case ".gif":
		encodeFunc = func(w io.Writer, img image.Image) error {
			return gif.Encode(w, img, nil)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", extension)
	}

	// Create the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Encode and save the image
	if err := encodeFunc(outFile, paddedImg); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

// getCachePath generates a unique cache path based on the input file.
func getCachePath(inputPath string) string {
	hash := md5.Sum([]byte(inputPath))
	return filepath.Join("cache", hex.EncodeToString(hash[:]))
}

// isCached checks if the image has already been processed and cached.
func isCached(cachePath, fileName string) bool {
	cacheFile := filepath.Join(cachePath, fileName+".cache")
	if _, err := os.Stat(cacheFile); err == nil {
		return true
	}
	return false
}

// updateCache marks an image as processed in the cache.
func updateCache(cachePath, fileName string) error {
	cacheFile := filepath.Join(cachePath, fileName+".cache")
	file, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
