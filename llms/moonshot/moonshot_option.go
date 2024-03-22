package moonshot

import "github.com/tmc/langchaingo/llms/moonshot/internal/moonshotclient"

const (
	tokenEnvVarName   = "OPENAI_API_KEY"  //nolint:gosec
	modelEnvVarName   = "OPENAI_MODEL"    //nolint:gosec
	baseURLEnvVarName = "OPENAI_BASE_URL" //nolint:gosec
)

type options struct {
	token      string
	model      string
	baseURL    string
	httpClient moonshotclient.Doer

	// required when APIType is APITypeAzure or APITypeAzureAD
	//apiVersion     string
	//embeddingModel string
}

type Option func(*options)

// WithToken passes the OpenAI API token to the client. If not set, the token
// is read from the OPENAI_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the OpenAI model to the client. If not set, the model
// is read from the OPENAI_MODEL environment variable.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithBaseURL passes the OpenAI base url to the client. If not set, the base url
// is read from the OPENAI_BASE_URL environment variable. If still not set in ENV
// VAR OPENAI_BASE_URL, then the default value is https://api.openai.com/v1 is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// WithHTTPClient allows setting a custom HTTP client. If not set, the default value
// is http.DefaultClient.
func WithHTTPClient(client moonshotclient.Doer) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}
