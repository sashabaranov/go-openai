package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// ---- These are the old structs and will likely be deprecated soon.
// ModerationRequest represents a request structure for moderation API
type ModerationRequest struct {
	Input string  `json:"input,omitempty"`
	Model *string `json:"model,omitempty"`
}

// Result represents one of possible moderation results
type Result struct {
	Categories     ResultCategories     `json:"categories"`
	CategoryScores ResultCategoryScores `json:"category_scores"`
	Flagged        int                  `json:"flagged"`
}

// ResultCategories represents Categories of Result
type ResultCategories struct {
	Hate            int `json:"hate"`
	HateThreatening int `json:"hate/threatening"`
	SelfHarm        int `json:"self-harm"`
	Sexual          int `json:"sexual"`
	SexualMinors    int `json:"sexual/minors"`
	Violence        int `json:"violence"`
	ViolenceGraphic int `json:"violence/graphic"`
}

// ResultCategoryScores represents CategoryScores of Result
type ResultCategoryScores struct {
	Hate            float32 `json:"hate"`
	HateThreatening float32 `json:"hate/threatening"`
	SelfHarm        float32 `json:"self-harm"`
	Sexual          float32 `json:"sexual"`
	SexualMinors    float32 `json:"sexual/minors"`
	Violence        float32 `json:"violence"`
	ViolenceGraphic float32 `json:"violence/graphic"`
}

// ModerationResponse represents a response structure for moderation API
type ModerationResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Results []Result `json:"results"`
}

// ModerationOld — perform a moderation api call over a string.
// Input can be an array or slice but a string will reduce the complexity.
func (c *Client) ModerationsOld(ctx context.Context, request ModerationRequest) (response ModerationResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.fullURL("/moderations"), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}

// ---- These are the new structs and will likely be used in starting Friday, August 5th 2022.
// ModerationRequestNew represents a request structure for moderation API
type ModerationRequestNew struct {
	Input string  `json:"input,omitempty"`
	Model *string `json:"model,omitempty"`
}

// ResultNew represents one of possible moderation NewResults
type ResultNew struct {
	Categories     ResultNewCategories     `json:"categories"`
	CategoryScores ResultNewCategoryScores `json:"category_scores"`
	Flagged        bool                    `json:"flagged"`
}

// ResultNewCategories represents Categories of ResultNew
type ResultNewCategories struct {
	Hate            bool `json:"hate"`
	HateThreatening bool `json:"hate/threatening"`
	SelfHarm        bool `json:"self-harm"`
	Sexual          bool `json:"sexual"`
	SexualMinors    bool `json:"sexual/minors"`
	Violence        bool `json:"violence"`
	ViolenceGraphic bool `json:"violence/graphic"`
}

// ResultNewCategoryScores represents CategoryScores of ResultNew
type ResultNewCategoryScores struct {
	Hate            float32 `json:"hate"`
	HateThreatening float32 `json:"hate/threatening"`
	SelfHarm        float32 `json:"self-harm"`
	Sexual          float32 `json:"sexual"`
	SexualMinors    float32 `json:"sexual/minors"`
	Violence        float32 `json:"violence"`
	ViolenceGraphic float32 `json:"violence/graphic"`
}

// ModerationResponseNew represents a response structure for moderation API
type ModerationResponseNew struct {
	ID         string      `json:"id"`
	Model      string      `json:"model"`
	NewResults []ResultNew `json:"NewResults"`
}

// ModerationsNew — perform a moderation api call over a string.
// Input can be an array or slice but a string will reduce the complexity.
func (c *Client) ModerationsNew(ctx context.Context, request ModerationRequestNew) (response ModerationResponseNew, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.fullURL("/moderations"), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}
