package openai

import (
	"context"
	"net/http"
)


type AssistantThreadRunStreamMessageDelta struct {
    Content []AssistantThreadRunStreamMessageDeltaContent   `jsno:"content"`
}

type AssistantThreadRunStreamResponse struct {
	ID                string                               `json:"id"`
	Object            string                               `json:"object"`
    /*
	Delta             AssistantThreadRunStreamMessageDelta `json:"delta"`
    */
}

type AssistantThreadRunStream struct {
	*streamReader[AssistantThreadRunStreamResponse]
}
