package moonshot

import (
	"errors"
	"github.com/tmc/langchaingo/llms/moonshot/internal/moonshotclient"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse              = errors.New("no response")
	ErrMissingToken               = errors.New("鉴权失败请确认")            //nolint:lll
	ErrExceededQuota              = errors.New("Quota 不够了，请联系管理员加量") //nolint:lll
	ErrRateLimitReached           = errors.New("您超速了，请稍后重试")         //nolint:lll
	ErrInvalidRequest             = errors.New("参数格式有误")             //nolint:lll
	ErrUnexpectedResponseLength   = errors.New("unexpected length of response")
	ErrMissingAzureEmbeddingModel = errors.New("embeddings model needs to be provided when using Azure API")
)

// newClient is wrapper for moonshotclient internal package.
func newClient(opts ...Option) (*moonshotclient.Client, error) {
	options := &options{
		token:      os.Getenv(tokenEnvVarName),
		model:      os.Getenv(modelEnvVarName),
		baseURL:    os.Getenv(baseURLEnvVarName),
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return moonshotclient.New(options.token, options.model, options.baseURL, options.httpClient)
}
