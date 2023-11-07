package openai_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestSpeechIntegration(t *testing.T) {
	client := openai.NewClient("sk-xyz")

	res, err := client.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
		Model: openai.TTS_MODEL_1,
		Input: "Hello!",
		Voice: openai.VOICE_ALLOY,
	})
	checks.NoError(t, err, "CreateSpeech error")
	defer res.Close()

	buf, err := io.ReadAll(res)
	checks.NoError(t, err, "ReadAll error")

	// save buf to file as mp3
	err = os.WriteFile("test.mp3", buf, 0644)
	checks.NoError(t, err, "Create error")
}
