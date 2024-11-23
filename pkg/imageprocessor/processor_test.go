package imageprocessor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/drewalth/logo-generator/pkg/logger"
)

func TestResizeAndSaveImage(t *testing.T) {
	// Initialize loggers
	infoLogger := logger.NewInfoLogger()
	errorLogger := logger.NewErrorLogger()

	// Open a sample image
	inputPath := "testdata/sample.png"
	outputDir := "testdata/output"
	dimensions := []Dimension{
		{Width: 100, Height: 100, Name: "test_resized.png"},
	}

	// Ensure output directory exists
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Process the image
	err := ProcessImage(ctx, inputPath, outputDir, dimensions, infoLogger, errorLogger)
	if err != nil {
		t.Fatalf("ProcessImage failed: %v", err)
	}

	// Check if the resized image exists
	resizedPath := filepath.Join(outputDir, "test_resized.png")
	if _, err := os.Stat(resizedPath); os.IsNotExist(err) {
		t.Errorf("Resized image does not exist: %s", resizedPath)
	}

	// Additional checks can be added to verify image properties
}

func TestLoadDimensions(t *testing.T) {
	configPath := "config/dimensions.json"
	dimensions, err := LoadDimensions(configPath)
	if err != nil {
		t.Fatalf("LoadDimensions failed: %v", err)
	}

	if len(dimensions) == 0 {
		t.Error("No dimensions loaded from config")
	}
}
