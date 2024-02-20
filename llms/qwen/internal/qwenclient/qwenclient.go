package qwenclient

import (
	"context"
	"errors"
	"net/http"
)

const (
	defaultBaseUrl      = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	defaultEmbeddingURL = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
)

var ErrEmptyResponse = errors.New("empty response")

type Client struct {
	apiKey          string
	baseURL         string
	model           string
	httpClient      Doer
	embeddingsModel string
	EnableSearch    bool
}

type Option func(client *Client) error
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(apiKey string,
	baseURL string,
	model string,
	httpClient Doer,
	embeddingsModel string,
	enableSearch bool,
	opts ...Option) (*Client, error) {
	c := &Client{
		apiKey:          apiKey,
		baseURL:         baseURL,
		model:           model,
		httpClient:      httpClient,
		embeddingsModel: embeddingsModel,
		EnableSearch:    enableSearch,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseUrl
	}
	if c.model == "" {
		c.model = defaultChatModel
	}
	if c.embeddingsModel == "" {
		c.embeddingsModel = defaultEmbeddingModel
	}
	return c, nil
}

// CreateCompletion 提供completion接口
func (c *Client) CreateCompletion(ctx context.Context, r *ChatRequestUser) (*ChatResponse, error) {
	resp, err := c.createCompletion(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Output.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp, nil
}

// CreateEmbedding 提供向量接口
func (c *Client) CreateEmbedding(ctx context.Context, r *EmbeddingPayload) ([][]float64, int, error) {
	if r.Model == "" {
		r.Model = defaultEmbeddingModel
	}

	resp, err := c.createEmbedding(ctx, r)
	if err != nil {
		return nil, 0, err
	}

	if len(resp.Output.Embeddings[0].Embedding) == 0 {
		return nil, 0, ErrEmptyResponse
	}

	embeddings := make([][]float64, 0, 1)
	for _, e := range resp.Output.Embeddings {
		embeddings = append(embeddings, e.Embedding)
	}
	return embeddings, resp.Usage.TotalTokens, nil
}

func (c *Client) CreateChat(ctx context.Context, r *ChatRequestUser) (*ChatResponse, error) {
	if r.Model == "" {
		if c.model == "" {
			r.Model = defaultChatModel
		} else {
			r.Model = c.model
		}
	}
	resp, err := c.createChat(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Output.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp, nil
}
