package qwen

import (
	"errors"
	"github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("缺少API ID 或者 API SECRET") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*qwenclient.Client, error) {
	options := &options{
		apiKey:     os.Getenv(apiKeyEnvName),
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.apiKey) == 0 {
		return nil, ErrMissingToken
	}
	return qwenclient.New(options.apiKey, options.baseURL, options.model,
		options.httpClient, options.embeddingModel, options.EnableSearch)
}
