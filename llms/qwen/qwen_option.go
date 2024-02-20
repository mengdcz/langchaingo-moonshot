package qwen

import "github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"

const (
	apiKeyEnvName = "DASHSCOPE_API_KEY"
)

type options struct {
	apiKey         string
	model          string
	baseURL        string
	httpClient     qwenclient.Doer
	embeddingModel string
	EnableSearch   bool
}

type Option func(*options)

type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

func WithApiKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
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
func WithHttpClient(httpClient qwenclient.Doer) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}
func WithEmbeddingModel(embeddingModel string) Option {
	return func(o *options) {
		o.embeddingModel = embeddingModel
	}
}
func WithEnableSearch(enableSearch bool) Option {
	return func(o *options) {
		o.EnableSearch = enableSearch
	}
}
