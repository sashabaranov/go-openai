package openai

import (
	"context"
	"fmt"
	"net/http"
)

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTuneRequest struct {
	TrainingFile                 string    `json:"training_file"`
	ValidationFile               string    `json:"validation_file,omitempty"`
	Model                        string    `json:"model,omitempty"`
	Epochs                       int       `json:"n_epochs,omitempty"`
	BatchSize                    int       `json:"batch_size,omitempty"`
	LearningRateMultiplier       float32   `json:"learning_rate_multiplier,omitempty"`
	PromptLossRate               float32   `json:"prompt_loss_rate,omitempty"`
	ComputeClassificationMetrics bool      `json:"compute_classification_metrics,omitempty"`
	ClassificationClasses        int       `json:"classification_n_classes,omitempty"`
	ClassificationPositiveClass  string    `json:"classification_positive_class,omitempty"`
	ClassificationBetas          []float32 `json:"classification_betas,omitempty"`
	Suffix                       string    `json:"suffix,omitempty"`
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTune struct {
	ID                string              `json:"id"`
	Object            string              `json:"object"`
	Model             string              `json:"model"`
	CreatedAt         int64               `json:"created_at"`
	FineTuneEventList []FineTuneEvent     `json:"events,omitempty"`
	FineTunedModel    string              `json:"fine_tuned_model"`
	HyperParams       FineTuneHyperParams `json:"hyperparams"`
	OrganizationID    string              `json:"organization_id"`
	ResultFiles       []File              `json:"result_files"`
	Status            string              `json:"status"`
	ValidationFiles   []File              `json:"validation_files"`
	TrainingFiles     []File              `json:"training_files"`
	UpdatedAt         int64               `json:"updated_at"`

	httpHeader
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTuneEvent struct {
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTuneHyperParams struct {
	BatchSize              int     `json:"batch_size"`
	LearningRateMultiplier float64 `json:"learning_rate_multiplier"`
	Epochs                 int     `json:"n_epochs"`
	PromptLossWeight       float64 `json:"prompt_loss_weight"`
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTuneList struct {
	Object string     `json:"object"`
	Data   []FineTune `json:"data"`

	httpHeader
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTuneEventList struct {
	Object string          `json:"object"`
	Data   []FineTuneEvent `json:"data"`

	httpHeader
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
type FineTuneDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
func (c *Client) CreateFineTune(ctx context.Context, request FineTuneRequest) (response FineTune, err error) {
	urlSuffix := "/fine-tunes"
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CancelFineTune cancel a fine-tune job.
// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
func (c *Client) CancelFineTune(ctx context.Context, fineTuneID string) (response FineTune, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/fine-tunes/"+fineTuneID+"/cancel")) //nolint:lll //this method is deprecated
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
func (c *Client) ListFineTunes(ctx context.Context) (response FineTuneList, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL("/fine-tunes"))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
func (c *Client) GetFineTune(ctx context.Context, fineTuneID string) (response FineTune, err error) {
	urlSuffix := fmt.Sprintf("/fine-tunes/%s", fineTuneID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
func (c *Client) DeleteFineTune(ctx context.Context, fineTuneID string) (response FineTuneDeleteResponse, err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL("/fine-tunes/"+fineTuneID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// Deprecated: On August 22nd, 2023, OpenAI announced the deprecation of the /v1/fine-tunes API.
// This API will be officially deprecated on January 4th, 2024.
// OpenAI recommends to migrate to the new fine tuning API implemented in fine_tuning_job.go.
func (c *Client) ListFineTuneEvents(ctx context.Context, fineTuneID string) (response FineTuneEventList, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL("/fine-tunes/"+fineTuneID+"/events"))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
