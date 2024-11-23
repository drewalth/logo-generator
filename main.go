package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/drewalth/logo-generator/pkg/imageprocessor"
	"github.com/drewalth/logo-generator/pkg/logger"
)

func main() {
	// Initialize loggers
	infoLogger := logger.NewInfoLogger()
	errorLogger := logger.NewErrorLogger()

	// Define command-line flags
	inputPath := flag.String("input", "", "Path to the input image")
	outputDir := flag.String("output", "output", "Directory to save resized images")
	configPath := flag.String("config", "config/dimensions.json", "Path to dimensions config file")
	timeout := flag.Duration("timeout", 2*time.Minute, "Timeout for image processing")

	flag.Parse()

	// Validate input
	if *inputPath == "" {
		flag.Usage()
		log.Fatal("Error: -input flag is required")
	}

	// Load dimensions from config
	dimensions, err := imageprocessor.LoadDimensions(*configPath)
	if err != nil {
		errorLogger.Fatalf("Failed to load dimensions: %v\n", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Process the image
	infoLogger.Println("Starting image processing...")
	err = imageprocessor.ProcessImage(ctx, *inputPath, *outputDir, dimensions, infoLogger, errorLogger)
	if err != nil {
		errorLogger.Fatalf("Image processing failed: %v\n", err)
	}

	infoLogger.Println("Image processing complete. Resized images saved to:", *outputDir)
}
