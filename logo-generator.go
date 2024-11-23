package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
)

// Dimensions to resize the image to.
// update the dimensions as needed
var dimensions = []struct {
	width  uint
	height uint
	name   string
}{
	{310, 310, "Square310x310Logo.png"},
	{284, 284, "Square284x284Logo.png"},
	{150, 150, "Square150x150Logo.png"},
	{142, 142, "Square142x142Logo.png"},
	{107, 107, "Square107x107Logo.png"},
	{89, 89, "Square89x89Logo.png"},
	{71, 71, "Square71x71Logo.png"},
	{44, 44, "Square44x44Logo.png"},
	{30, 30, "Square30x30Logo.png"},
	{512, 512, "icon.png"},
	{512, 512, "icon.icns"},
	{256, 256, "icon.ico"},
	{256, 256, "128x128@2x.png"},
	{50, 50, "StoreLogo.png"},
	{128, 128, "128x128.png"},
	{32, 32, "32x32.png"},
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run logo-generator.go <path_to_image>")
	}

	imagePath := os.Args[1]
	outputDir := "output"

	if err := processImage(imagePath, outputDir); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Println("Image processing complete. Resized images saved to:", outputDir)
}

// processImage reads the input image, validates its format and size,
// and generates resized images in specified dimensions.
func processImage(inputPath, outputDir string) error {
	// Open the input image file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// Decode the PNG image
	srcImg, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Validate the image dimensions
	bounds := srcImg.Bounds()
	if bounds.Dx() != 1080 || bounds.Dy() != 1080 {
		return fmt.Errorf("image dimensions must be 1080x1080, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate resized images
	for _, dim := range dimensions {
		if err := resizeAndSaveRGBAImage(srcImg, dim.width, dim.height, filepath.Join(outputDir, dim.name)); err != nil {
			return fmt.Errorf("failed to create resized image %s: %v", dim.name, err)
		}
	}

	return nil
}

// resizeAndSaveRGBAImage resizes the source image to the specified dimensions,
// converts it to RGBA format, and saves it to the specified output path.
func resizeAndSaveRGBAImage(src image.Image, width, height uint, outputPath string) error {
	// Resize the image to the specified dimensions
	resizedImg := resize.Resize(width, height, src, resize.Lanczos3)

	// Convert the resized image to RGBA format
	rgbaImg := image.NewRGBA(resizedImg.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), resizedImg, image.Point{}, draw.Over)

	// Remove or comment out the applyAlpha function to preserve original alpha
	// applyAlpha(rgbaImg)

	// Save the resized RGBA image to the specified file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Encode and save the resized RGBA image as PNG
	if err := png.Encode(outFile, rgbaImg); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

// applyAlpha ensures the alpha channel is properly set for the RGBA image.
// In this example, it retains transparency if present or applies a full-opacity alpha channel.
func applyAlpha(img *image.RGBA) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// If alpha is missing, set it to full opacity (255)
			if a == 0 {
				img.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255})
			}
		}
	}
}
