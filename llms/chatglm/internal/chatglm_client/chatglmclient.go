package chatglm_client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// chatglm  ChatGLM-Pro  ChatGLM-Std 上下文长度 8k
	//  https://open.bigmodel.cn/api/paas/v3/model-api/{model}/{invoke_method}
	defaultBaseURL = "https://open.bigmodel.cn/api/paas/v3/model-api/%s/%s"
)

var ErrEmptyResponse = errors.New("empty response")

type Client struct {
	id              string
	secret          string
	baseURL         string
	model           string
	token           string
	tokenExpireTime int64
	httpClient      Doer
	embeddingsModel string
	cache           Cache
	EnableSearch    bool
	SearchQuery     string
}

type Option func(client *Client) error

// Cache 公共缓存
type Cache interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Expire(key string, seconds int) error
}

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(id string,
	secret string,
	model string,
	baseURL string,
	token string,
	tokenExpireTime int64,
	httpClient Doer,
	embeddingsModel string,
	cache Cache,
	enableSearch bool,
	searchQuery string,
	opts ...Option) (*Client, error) {
	c := &Client{
		id:              id,
		secret:          secret,
		baseURL:         baseURL,
		model:           model,
		token:           token,
		tokenExpireTime: tokenExpireTime,
		httpClient:      httpClient,
		embeddingsModel: embeddingsModel,
		cache:           cache,
		EnableSearch:    enableSearch,
		SearchQuery:     searchQuery,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
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
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	resp, err := c.createCompletion(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 || !resp.Success || len(resp.Data.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return &Completion{
		Text:  resp.Data.Choices[0].Content,
		Usage: resp.Data.Usage,
	}, nil
}

type EmbeddingRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	RequestId string `json:"request_id"`
}

// CreateEmbedding 提供向量接口
func (c *Client) CreateEmbedding(ctx context.Context, r *EmbeddingRequest) ([]float64, Usage, error) {
	if r.Model == "" {
		r.Model = defaultEmbeddingModel
	}

	resp, err := c.createEmbedding(ctx, &embeddingPayload{
		Prompt:    r.Prompt,
		RequestId: r.RequestId,
	})
	if err != nil {
		return nil, Usage{}, err
	}

	if resp.Code != 200 || !resp.Success || len(resp.Data.Embedding) == 0 {
		return nil, Usage{}, ErrEmptyResponse
	}

	embeddings := make([]float64, 1024)
	copy(embeddings, resp.Data.Embedding)
	return embeddings, resp.Data.Usage, nil
}

// CreateChat 提供chat接口
func (c *Client) CreateChat(ctx context.Context, r *ChatRequest) (*ChatResponse, error) {
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
	if resp.Code != 200 || !resp.Success || len(resp.Data.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp, nil
}

// 设置权限
func (c *Client) setHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.getAuthorization())
}

func (c *Client) buildURL(model string, method string) string {
	return fmt.Sprintf(c.baseURL, model, method)
}

func (c *Client) getAuthorization() string {
	if c.token != "" && c.tokenExpireTime > time.Now().Unix() {
		return c.token
	}
	// 过期重置
	c.token = ""
	c.tokenExpireTime = 0
	var err error
	if c.cache != nil {
		c.token, err = c.cache.Get(c.getKey())
	}
	if c.token != "" && err != nil {
		return c.token
	}
	// 生成token
	timestamp := time.Now().UnixMilli()
	exp := timestamp + 3600*1000
	c.token, err = CreateToken([]byte(c.secret), map[string]any{
		"api_key":   c.id,
		"exp":       exp,
		"timestamp": timestamp,
	})
	if err != nil {
		c.token = ""
		return c.token
	}
	c.tokenExpireTime = exp
	if c.cache != nil {
		// 错误丢弃
		c.cache.Set(c.getKey(), c.token)
		c.cache.Expire(c.getKey(), 3600-600)
	}
	return c.token
}

func (c *Client) getKey() string {
	return "kpai:chatglm:" + c.secret
}
