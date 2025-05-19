package openai

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
)

type Marshaller interface {
	Marshal(value any) ([]byte, error)
}

type JSONMarshaller struct{}

func (jm *JSONMarshaller) Marshal(value any) ([]byte, error) {
	originalBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	// Check if the value implements the GetExtraFields interface
	getExtraFieldsBody, ok := value.(interface {
		GetExtraFields() map[string]any
	})
	if !ok {
		// If not, return the original bytes
		return originalBytes, nil
	}
	extraFields := getExtraFieldsBody.GetExtraFields()
	if len(extraFields) == 0 {
		// If there are no extra fields, return the original bytes
		return originalBytes, nil
	}
	patchBytes, err := json.Marshal(extraFields)
	if err != nil {
		return nil, fmt.Errorf("Marshal extraFields(%+v) err: %w", extraFields, err)
	}
	finalBytes, err := jsonpatch.MergePatch(originalBytes, patchBytes)
	if err != nil {
		return nil, fmt.Errorf("MergePatch originalBytes(%s) patchBytes(%s) err: %w", originalBytes, patchBytes, err)
	}
	return finalBytes, nil
}
