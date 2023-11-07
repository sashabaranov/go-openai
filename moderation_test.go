package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

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
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/moderations", handleModerationEndpoint)
	_, err := client.Moderations(context.Background(), ModerationRequest{
		Model: ModerationTextStable,
		Input: "I want to kill them.",
	})
	checks.NoError(t, err, "Moderation error")
}

// TestModerationsWithIncorrectModel Tests passing valid and invalid models to moderations endpoint.
func TestModerationsWithDifferentModelOptions(t *testing.T) {
	var modelOptions []struct {
		model  string
		expect error
	}
	modelOptions = append(modelOptions,
		getModerationModelTestOption(GPT3Dot5Turbo, ErrModerationInvalidModel),
		getModerationModelTestOption(ModerationTextStable, nil),
		getModerationModelTestOption(ModerationTextLatest, nil),
		getModerationModelTestOption("", nil),
	)
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/moderations", handleModerationEndpoint)
	for _, modelTest := range modelOptions {
		_, err := client.Moderations(context.Background(), ModerationRequest{
			Model: modelTest.model,
			Input: "I want to kill them.",
		})
		checks.ErrorIs(t, err, modelTest.expect,
			fmt.Sprintf("Moderations(..) expects err: %v, actual err:%v", modelTest.expect, err))
	}
}

func getModerationModelTestOption(model string, expect error) struct {
	model  string
	expect error
} {
	return struct {
		model  string
		expect error
	}{model: model, expect: expect}
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
		Model: moderationReq.Model,
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
