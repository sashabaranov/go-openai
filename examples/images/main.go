package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

// saveImageToFile saves image data to a file in the specified directory
// Returns the full path where the image was saved or an error
func saveImageToFile(imageData []byte, outputDir, filename string) (string, error) {
	// Create output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write the image data to a file
	outputPath := filepath.Join(outputDir, filename)
	err := os.WriteFile(outputPath, imageData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write image file: %w", err)
	}

	return outputPath, nil
}

func main() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	res, err := client.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:            "Parrot on a skateboard performing a trick. Large bold text \"SKATE MASTER\" banner at the bottom of the image. Cartoon style, natural light, high detail",
			Background:        openai.CreateImageBackgroundTransparent,
			Model:             openai.CreateImageModelGptImage1,
			Size:              openai.CreateImageSize1024x1024,
			N:                 1,
			Quality:           openai.CreateImageQualityLow,
			OutputCompression: 100,
			OutputFormat:      openai.CreateImageOutputFormatJPEG,
			// ResponseFormat:    openai.CreateImageResponseFormatB64JSON,
		},
	)

	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	// Decode the base64 data
	imageBase64 := res.Data[0].B64JSON
	fmt.Printf("Base64 data: %s\n", imageBase64)

	imageBytes, err := base64.StdEncoding.DecodeString(res.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	// Save the image using the new function
	outputPath, err := saveImageToFile(imageBytes, "examples/images", "generated_image.jpg")
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return
	}

	fmt.Printf("Image saved to: %s\n", outputPath)
}
