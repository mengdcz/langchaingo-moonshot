package chatglm

import (
	"errors"
	"github.com/tmc/langchaingo/llms/chatglm/internal/chatglm_client"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("缺少API ID 或者 API SECRET") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*chatglm_client.Client, error) {
	options := &options{
		id:         os.Getenv(apiIdEnvName),
		secret:     os.Getenv(apiSecretEnvName),
		baseURL:    os.Getenv(baseUrlEnvName),
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.id) == 0 || len(options.secret) == 0 {
		return nil, ErrMissingToken
	}
	return chatglm_client.New(options.id, options.secret, options.model, options.baseURL, options.token, options.tokenExpireTime,
		options.httpClient, options.embeddingModel, options.cache, options.enableSearch, options.searchQuery)
}
