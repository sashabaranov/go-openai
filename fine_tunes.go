package openai

import (
	"context"
	"fmt"
	"net/http"
)

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
}

type FineTuneEvent struct {
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type FineTuneHyperParams struct {
	BatchSize              int     `json:"batch_size"`
	LearningRateMultiplier float64 `json:"learning_rate_multiplier"`
	Epochs                 int     `json:"n_epochs"`
	PromptLossWeight       float64 `json:"prompt_loss_weight"`
}

type FineTuneList struct {
	Object string     `json:"object"`
	Data   []FineTune `json:"data"`
}
type FineTuneEventList struct {
	Object string          `json:"object"`
	Data   []FineTuneEvent `json:"data"`
}

type FineTuneDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

func (c *Client) CreateFineTune(ctx context.Context, request FineTuneRequest) (response FineTune, err error) {
	urlSuffix := "/fine-tunes"
	req, err := c.requestBuilder.Build(ctx, http.MethodPost, c.fullURL(urlSuffix), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CancelFineTune cancel a fine-tune job.
func (c *Client) CancelFineTune(ctx context.Context, fineTuneID string) (response FineTune, err error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodPost, c.fullURL("/fine-tunes/"+fineTuneID+"/cancel"), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

func (c *Client) ListFineTunes(ctx context.Context) (response FineTuneList, err error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, c.fullURL("/fine-tunes"), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

func (c *Client) GetFineTune(ctx context.Context, fineTuneID string) (response FineTune, err error) {
	urlSuffix := fmt.Sprintf("/fine-tunes/%s", fineTuneID)
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, c.fullURL(urlSuffix), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

func (c *Client) DeleteFineTune(ctx context.Context, fineTuneID string) (response FineTuneDeleteResponse, err error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodDelete, c.fullURL("/fine-tunes/"+fineTuneID), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

func (c *Client) ListFineTuneEvents(ctx context.Context, fineTuneID string) (response FineTuneEventList, err error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, c.fullURL("/fine-tunes/"+fineTuneID+"/events"), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
