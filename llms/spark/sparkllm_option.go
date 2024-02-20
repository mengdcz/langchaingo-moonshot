package spark

import (
	"github.com/tmc/langchaingo/llms/spark/internal/sparkclient"
)

const (
	idEnvVarName     = "SPARK_APP_ID"
	secretEnvVarName = "SPARK_APP_SECRET"
	keyEnvVarname    = "SPARK_APP_KEY"

	modelEnvVarName   = "OPENAI_MODEL"
	baseURLEnvVarName = "OPENAI_BASE_URL"
)

type options struct {
	id             string
	secret         string
	key            string
	model          string
	baseURL        string
	embeddingModel string
	httpClient     sparkclient.Doer
}

type Option func(*options)

func WithId(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

func WithSecret(secret string) Option {
	return func(o *options) {
		o.secret = secret
	}
}

func WithKey(key string) Option {
	return func(o *options) {
		o.key = key
	}
}

func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

func WithEmbeddingModel(embeddingModel string) Option {
	return func(o *options) {
		o.embeddingModel = embeddingModel
	}
}

func WithHttpClient(httpClient sparkclient.Doer) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}
