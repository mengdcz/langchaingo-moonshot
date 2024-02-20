package minimax

import (
	"errors"
	"fmt"
	minimaxclient2 "github.com/tmc/langchaingo/llms/minimax/minimaxclient"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("缺少GROUP ID 或者 API KEY") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*minimaxclient2.Client, error) {
	options := &options{
		groupId:        os.Getenv(groupIdEnvVarName),
		apiKey:         os.Getenv(apiKeyEnvVarName),
		baseUrl:        os.Getenv(baseURLEnvVarName),
		httpClient:     http.DefaultClient,
		embeddingModel: "",
		model:          "",
	}

	fmt.Println(options.groupId)

	for _, opt := range opts {
		opt(options)
	}

	if options.model == "" {
		options.model = defaultModel
	}

	if options.embeddingModel == "" {
		options.embeddingModel = defaultEmbeddingModel
	}

	return minimaxclient2.NewClient(minimaxclient2.WithGroupId(options.groupId),
		minimaxclient2.WithApiKey(options.apiKey),
		minimaxclient2.WithBaseUrl(options.baseUrl),
		minimaxclient2.WithHttpClient(options.httpClient),
		minimaxclient2.WithModel(options.model),
		minimaxclient2.WithEmbeddingsModel(options.embeddingModel),
	)
}
