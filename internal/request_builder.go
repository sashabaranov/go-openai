package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Method       string
	URL          string
	Body         any
	Header       http.Header
	ExtraBody    map[string]any
	ExtraHeaders map[string]string
	ExtraQuery   map[string]string
}

type RequestBuilder interface {
	Build(ctx context.Context, request *Request) (*http.Request, error)
}

type HTTPRequestBuilder struct {
	marshaller Marshaller
}

func NewRequestBuilder() *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		marshaller: &JSONMarshaller{},
	}
}

func (b *HTTPRequestBuilder) Build(
	ctx context.Context,
	request *Request,
) (req *http.Request, err error) {
	var bodyReader io.Reader
	if request.Body != nil {
		if v, ok := request.Body.(io.Reader); ok {
			bodyReader = v
		} else {
			var reqBytes []byte
			reqBytes, err = b.marshaller.Marshal(request.Body)
			if err != nil {
				return
			}

			if request.ExtraBody != nil {
				rawMap := make(map[string]any)
				err = b.marshaller.Unmarshal(reqBytes, &rawMap)
				if err != nil {
					return
				}

				for k, v := range request.ExtraBody {
					rawMap[k] = v
				}
				reqBytes, err = b.marshaller.Marshal(rawMap)
				if err != nil {
					return
				}
			}

			bodyReader = bytes.NewBuffer(reqBytes)
		}
	}

	requestUrl := request.URL
	if request.ExtraQuery != nil {
		for k, v := range request.ExtraQuery {
			requestUrl = fmt.Sprintf("%s&%s=%s", requestUrl, k, url.QueryEscape(v))
		}
	}

	requestUrl = strings.TrimSuffix(requestUrl, "&")

	req, err = http.NewRequestWithContext(ctx, request.Method, requestUrl, bodyReader)
	if err != nil {
		return
	}
	if request.Header != nil {
		req.Header = request.Header
	} else {
		req.Header = make(http.Header)
	}

	for k, v := range request.ExtraHeaders {
		req.Header.Set(k, v)
	}

	return
}
