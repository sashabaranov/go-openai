package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please provide a filename to convert to text")
		return
	}
	if _, err := os.Stat(os.Args[1]); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("file %s does not exist\n", os.Args[1])
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:                   openai.Whisper1,
			FilePath:                os.Args[1],
			Format:                  openai.AudioResponseFormatVerboseJSON,
			Timestamp_Granularities: openai.TimestampGranularitiesWord, // Timestamp granularities are only supported with response_format=verbose_json
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}

	jsonOutput, err := json.Marshal(&resp.Words)
	if err != nil {
		fmt.Printf("Unmarshal JSON error: %v\n", err)
		return
	}

	fmt.Println(string(jsonOutput))
	fmt.Println(resp.Text)
}
