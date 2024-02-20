package spark

import (
	"errors"
	"github.com/tmc/langchaingo/llms/spark/internal/sparkclient"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse              = errors.New("no response")
	ErrMissingToken               = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable") //nolint:lll
	ErrMissingAzureEmbeddingModel = errors.New("embeddings model needs to be provided when using Azure API")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*sparkclient.Client, error) {
	options := &options{
		id:         os.Getenv(idEnvVarName),
		secret:     os.Getenv(secretEnvVarName),
		key:        os.Getenv(keyEnvVarname),
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}
	return sparkclient.New(options.id, options.secret, options.key, options.model, options.baseURL, options.embeddingModel, options.httpClient)
}
