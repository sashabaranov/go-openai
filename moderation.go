package openai

import (
	"context"
	"net/http"
)

// The moderation endpoint is a tool you can use to check whether content complies with OpenAI's usage policies.
// Developers can thus identify content that our usage policies prohibits and take action, for instance by filtering it.

// The default is text-moderation-latest which will be automatically upgraded over time.
// This ensures you are always using our most accurate model.
// If you use text-moderation-stable, we will provide advanced notice before updating the model.
// Accuracy of text-moderation-stable may be slightly lower than for text-moderation-latest.
const (
	ModerationTextStable = "text-moderation-stable"
	ModerationTextLatest = "text-moderation-latest"
	ModerationText001    = "text-moderation-001"
)

// ModerationRequest represents a request structure for moderation API.
type ModerationRequest struct {
	Input string `json:"input,omitempty"`
	Model string `json:"model,omitempty"`
}

// Result represents one of possible moderation results.
type Result struct {
	Categories     ResultCategories     `json:"categories"`
	CategoryScores ResultCategoryScores `json:"category_scores"`
	Flagged        bool                 `json:"flagged"`
}

// ResultCategories represents Categories of Result.
type ResultCategories struct {
	Hate            bool `json:"hate"`
	HateThreatening bool `json:"hate/threatening"`
	SelfHarm        bool `json:"self-harm"`
	Sexual          bool `json:"sexual"`
	SexualMinors    bool `json:"sexual/minors"`
	Violence        bool `json:"violence"`
	ViolenceGraphic bool `json:"violence/graphic"`
}

// ResultCategoryScores represents CategoryScores of Result.
type ResultCategoryScores struct {
	Hate            float32 `json:"hate"`
	HateThreatening float32 `json:"hate/threatening"`
	SelfHarm        float32 `json:"self-harm"`
	Sexual          float32 `json:"sexual"`
	SexualMinors    float32 `json:"sexual/minors"`
	Violence        float32 `json:"violence"`
	ViolenceGraphic float32 `json:"violence/graphic"`
}

// ModerationResponse represents a response structure for moderation API.
type ModerationResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Results []Result `json:"results"`
}

// Moderations — perform a moderation api call over a string.
// Input can be an array or slice but a string will reduce the complexity.
func (c *Client) Moderations(ctx context.Context, request ModerationRequest) (response ModerationResponse, err error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodPost, c.fullURL("/moderations", request.Model), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
