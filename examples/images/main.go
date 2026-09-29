package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}
	client := openai.NewClient(key)
	response, err := client.CreateImage(context.Background(), openai.ImageRequest{
		Model:        openai.CreateImageModelGptImage2Dot5Flare,
		Prompt:       "Parrot on a skateboard performs a trick, cartoon style, natural light, high detail",
		Size:         openai.CreateImageSize1024x1024,
		Quality:      openai.CreateImageQualityLow,
		OutputFormat: openai.CreateImageOutputFormatPNG,
		N:            1,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(response.Data) == 0 || response.Data[0].B64JSON == "" {
		log.Fatal("no image data returned")
	}
	data, err := base64.StdEncoding.DecodeString(response.Data[0].B64JSON)
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile("image.png", data, 0o600); err != nil {
		log.Fatal(err)
	}
	log.Println("Saved image.png")
}
