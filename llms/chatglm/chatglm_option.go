package chatglm

import "github.com/tmc/langchaingo/llms/chatglm/internal/chatglm_client"

const (
	apiIdEnvName     = "CHATGLM_API_ID"
	apiSecretEnvName = "CHATGLM_API_SECRET"
	baseUrlEnvName   = "CHATGLM_BASE_URL"
)

type options struct {
	id              string
	secret          string
	model           string
	baseURL         string
	token           string
	tokenExpireTime int64
	httpClient      chatglm_client.Doer
	embeddingModel  string
	cache           chatglm_client.Cache
	enableSearch    bool
	searchQuery     string
}

type Option func(*options)

type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

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
func WithToken(token string) Option {
	return func(o *options) {
		o.token = token
	}
}
func WithTokenExpireTime(tokenExpireTime int64) Option {
	return func(o *options) {
		o.tokenExpireTime = tokenExpireTime
	}
}
func WithHttpClient(httpClient chatglm_client.Doer) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}
func WithEmbeddingModel(embeddingModel string) Option {
	return func(o *options) {
		o.embeddingModel = embeddingModel
	}
}
func WithCache(cache chatglm_client.Cache) Option {
	return func(o *options) {
		o.cache = cache
	}
}
func WithEnableSearch(enableSearch bool) Option {
	return func(o *options) {
		o.enableSearch = enableSearch
	}
}
func WithSearchQuery(searchQuery string) Option {
	return func(o *options) {
		o.searchQuery = searchQuery
	}
}
