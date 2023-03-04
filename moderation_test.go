package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestModeration Tests the moderations endpoint of the API using the mocked server.
func TestModerations(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/moderations", handleModerationEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	// create an edit request
	model := "text-moderation-stable"
	moderationReq := ModerationRequest{
		Model: &model,
		Input: "I want to kill them.",
	}
	_, err = client.Moderations(ctx, moderationReq)
	if err != nil {
		t.Fatalf("Moderation error: %v", err)
	}
}

// handleModerationEndpoint Handles the moderation endpoint by the test server.
func handleModerationEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var moderationReq ModerationRequest
	if moderationReq, err = getModerationBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	resCat := ResultCategories{}
	resCatScore := ResultCategoryScores{}
	switch {
	case strings.Contains(moderationReq.Input, "kill"):
		resCat = ResultCategories{Violence: true}
		resCatScore = ResultCategoryScores{Violence: 1}
	case strings.Contains(moderationReq.Input, "hate"):
		resCat = ResultCategories{Hate: true}
		resCatScore = ResultCategoryScores{Hate: 1}
	case strings.Contains(moderationReq.Input, "suicide"):
		resCat = ResultCategories{SelfHarm: true}
		resCatScore = ResultCategoryScores{SelfHarm: 1}
	case strings.Contains(moderationReq.Input, "porn"):
		resCat = ResultCategories{Sexual: true}
		resCatScore = ResultCategoryScores{Sexual: 1}
	}

	result := Result{Categories: resCat, CategoryScores: resCatScore, Flagged: true}

	res := ModerationResponse{
		ID:    strconv.Itoa(int(time.Now().Unix())),
		Model: *moderationReq.Model,
	}
	res.Results = append(res.Results, result)

	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getModerationBody Returns the body of the request to do a moderation.
func getModerationBody(r *http.Request) (ModerationRequest, error) {
	moderation := ModerationRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return ModerationRequest{}, err
	}
	err = json.Unmarshal(reqBody, &moderation)
	if err != nil {
		return ModerationRequest{}, err
	}
	return moderation, nil
}
