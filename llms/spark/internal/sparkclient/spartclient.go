package sparkclient

import (
	"context"
	"errors"
	"net/http"
)

const (
	// 星火大模型V1.5请求地址，对应的domain参数为general：
	defaultBaseUrl1 = "wss://spark-api.xf-yun.com/v1.1/chat"
	// 星火大模型V2请求地址，对应的domain参数为generalv2
	defaultBaseUrl2 = "wss://spark-api.xf-yun.com/v2.1/chat"
	// 星火大模型V3请求地址，对应的domain参数为generalv3
	defaultBaseUrl30 = "wss://spark-api.xf-yun.com/v3.1/chat"
	// embedding
	defaultBaseUrl3 = "http://knowledge-retrieval.cn-huabei-1.xf-yun.com/v1/aiui/embedding/query"

	modelName15 = "spark1.5"
	modelName20 = "spark2.0"
	modelName30 = "spark3.0"
)

var ErrEmptyResponse = errors.New("empty response")

type Client struct {
	appId           string
	appSecret       string
	apiKey          string
	model           string
	baseUrl         string
	embeddingsModel string
	httpClient      Doer
}

type Option func(*Client) error

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(appId, appSecret, apiKey, model, baseUrl, embeddingModel string, httpClient Doer, opts ...Option) (*Client, error) {
	c := &Client{
		appId:           appId,
		appSecret:       appSecret,
		apiKey:          apiKey,
		model:           model,
		baseUrl:         baseUrl,
		embeddingsModel: embeddingModel,
		httpClient:      httpClient,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.model == "" {
		c.model = modelName20
	}
	switch c.model {
	case modelName15:
		c.baseUrl = defaultBaseUrl1
	case modelName20:
		c.baseUrl = defaultBaseUrl2
	case modelName30:
		c.baseUrl = defaultBaseUrl30
	case defaultEmbeddingModel:
		c.baseUrl = defaultBaseUrl3
	}
	return c, nil
}

// CreateCompletion 提供completion接口
func (c *Client) CreateCompletion(ctx context.Context, r *ChatRequestUser) (*Completion, error) {
	resp, err := c.createCompletion(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Text) == 0 {
		return nil, ErrEmptyResponse
	}
	return &Completion{
		Text:  resp.Text,
		Usage: resp.Usage,
	}, nil
}

func (c *Client) CreateEmbedding(ctx context.Context, r *EmbeddingPayloadUser) ([]float64, error) {
	resp, err := c.createEmbedding(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload.Vector) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp.Payload.Vector, nil
}

func (c *Client) CreateChat(ctx context.Context, r *ChatRequestUser) (*ChatResponse, error) {
	resp, err := c.createChat(ctx, r)
	if err != nil {
		return nil, err
	}
	//if len(resp.Text) == 0 {
	//	return nil, ErrEmptyResponse
	//}
	return resp, nil
}
