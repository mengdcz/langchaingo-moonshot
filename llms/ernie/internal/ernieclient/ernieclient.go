package ernieclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"log"
	"net/http"
	"strings"
)

var (
	ErrNotSetAuth      = errors.New("both accessToken and apiKey secretKey are not set")
	ErrCompletionCode  = errors.New("completion API returned unexpected status code")
	ErrAccessTokenCode = errors.New("get access_token API returned unexpected status code")
	ErrEmbeddingCode   = errors.New("embedding API returned unexpected status code")
)

// Client is a client for the ERNIE API.
type Client struct {
	apiKey      string
	secretKey   string
	accessToken string
	httpClient  Doer
	cache       Cache
}

// Cache 公共缓存
type Cache interface {
	Set(key string, value string, second int) error
	Get(key string) (string, error)
}

// ModelPath ERNIE API URL path suffix distinguish models.
type ModelPath string

// DefaultCompletionModelPath default model.
const (
	DefaultCompletionModelPath = "completions"
	tryPeriod                  = 3 // minutes
)

// Option is an option for the ERNIE client.
type Option func(*Client) error

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Message struct {
	Role         string               `json:"role,omitempty"`
	Content      string               `json:"content,omitempty"`
	Name         string               `json:"name,omitempty"`
	FunctionCall *schema.FunctionCall `json:"function_call,omitempty"`
}

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Messages      []*Message                                    `json:"messages"`
	Temperature   float64                                       `json:"temperature,omitempty"`
	TopP          float64                                       `json:"top_p,omitempty"`
	PenaltyScore  float64                                       `json:"penalty_score,omitempty"`
	Stream        bool                                          `json:"stream,omitempty"`
	System        string                                        `json:"system,omitempty"`
	UserID        string                                        `json:"user_id,omitempty"`
	Functions     []llms.FunctionDefinition                     `json:"functions,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type Example Message

// Completion is a completion.
type Completion struct {
	ID               string              `json:"id"`
	Object           string              `json:"object"`
	Created          int                 `json:"created"`
	SentenceID       int                 `json:"sentence_id"`
	IsEnd            bool                `json:"is_end"`
	IsTruncated      bool                `json:"is_truncated"`
	Result           string              `json:"result"`
	NeedClearHistory bool                `json:"need_clear_history"`
	BanRound         int                 `json:"ban_round,omitempty"`
	Usage            Usage               `json:"usage"`
	FunctionCall     schema.FunctionCall `json:"function_call,omitempty"`
	// for error
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type Usage struct {
	PromptTokens     int           `json:"prompt_tokens,omitempty"`
	CompletionTokens int           `json:"completion_tokens,omitempty"`
	TotalTokens      int           `json:"total_tokens,omitempty"`
	PluginUsage      []PluginUsage `json:"plugin_usage,omitempty"`
}

type PluginUsage struct {
	Name           string `json:"name"`            //plugin名称，chatFile：chatfile插件消耗的tokens
	ParseTokens    int    `json:"parse_tokens"`    //解析文档tokens
	AbstractTokens int    `json:"abstract_tokens"` //摘要文档tokens
	SearchTokens   int    `json:"search_tokens"`   // 检索文档tokens
	TotalTokens    int    `json:"total_tokens"`    // 总tokens
}

type EmbeddingResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Data    []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage Usage `json:"usage"`
	// for error
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type authResponse struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	Scope         string `json:"scope"`
	SessionSecret string `json:"session_secret"`
}

// WithHTTPClient allows setting a custom HTTP client.
func WithHTTPClient(client Doer) Option {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// WithAKSK allows setting apiKey, secretKey.
func WithAKSK(apiKey, secretKey string) Option {
	return func(c *Client) error {
		c.apiKey = apiKey
		c.secretKey = secretKey
		return nil
	}
}

// Usually used for dev, Prod env recommend use WithAKSK.
func WithAccessToken(accessToken string) Option {
	return func(c *Client) error {
		c.accessToken = accessToken
		return nil
	}
}
func WithCache(cache Cache) Option {
	return func(c *Client) error {
		c.cache = cache
		return nil
	}
}

// New returns a new ERNIE client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.accessToken == "" && (c.apiKey == "" || c.secretKey == "") {
		return nil, ErrNotSetAuth
	}

	if c.apiKey != "" && c.secretKey != "" {
		err := autoRefresh(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func autoRefresh(c *Client) error {
	key := "langchain:ernieclient:" + c.apiKey
	var token string
	var err error
	if c.cache != nil {
		token, err = c.cache.Get(key)
		if err != nil {
			fmt.Println("1 get token from cache:", err.Error())
		}
		fmt.Println("2 get token from cache:", token)
	}
	//token = "24.af3b014b45adcddefb038109407ebf15.2592000.1703232487.282335-38624030"
	//token = "24.f6de479d0bd5b90449482e7b1e05248f.2592000.1703232748.282335-38624030"
	//token = ""
	if token != "" {
		c.accessToken = token
		fmt.Println("return")
		return nil
	}

	authResp, err := c.getAccessToken(context.Background())
	if err != nil {
		return err
	}
	fmt.Println(authResp.ExpiresIn)
	c.accessToken = authResp.AccessToken
	if c.cache != nil {
		//err = c.cache.Set(context.Background(), key, c.accessToken, time.Duration(authResp.ExpiresIn-1800))
		// 文心token默认30天， 缓存28天
		err = c.cache.Set(key, c.accessToken, authResp.ExpiresIn-3600*24*2)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	//go func() { // 30 day expiration, auto refresh access token per 10 days
	//	for {
	//		authResp, err := c.getAccessToken(context.Background())
	//		if err != nil {
	//			time.Sleep(tryPeriod * time.Minute) // try
	//			continue
	//		}
	//		c.accessToken = authResp.AccessToken
	//		time.Sleep(10 * 24 * time.Hour)
	//	}
	//}()
	return nil
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, modelPath ModelPath, r *CompletionRequest) (*Completion, error) {
	if modelPath == "" {
		modelPath = DefaultCompletionModelPath
	}

	url := "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/" + string(modelPath) +
		"?access_token=" + c.accessToken
	body, e := json.Marshal(r)
	if e != nil {
		return nil, e
	}
	fmt.Println(url)
	fmt.Println(string(body))
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrCompletionCode, resp.StatusCode)
	}

	if r.Stream {
		return parseStreamingCompletionResponse(ctx, resp, r)
	}

	var response Completion
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

// CreateEmbedding use ernie Embedding-V1.
func (c *Client) CreateEmbedding(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	url := "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1?access_token=" +
		c.accessToken

	payload := make(map[string]any)
	payload["input"] = texts

	body, e := json.Marshal(payload)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrEmbeddingCode, resp.StatusCode)
	}

	var response EmbeddingResponse
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

// accessToken 30 day expiration https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Ilkkrb0i5
func (c *Client) getAccessToken(ctx context.Context) (*authResponse, error) {
	url := fmt.Sprintf(
		"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%v&client_secret=%v",
		c.apiKey, c.secretKey)

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("")))
	if e != nil {
		return nil, e
	}

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrAccessTokenCode, resp.StatusCode)
	}

	var response authResponse
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

func parseStreamingCompletionResponse(ctx context.Context, resp *http.Response, req *CompletionRequest) (*Completion, error) { // nolint:lll
	scanner := bufio.NewScanner(resp.Body)
	responseChan := make(chan *Completion)
	go func() {
		defer close(responseChan)
		dataPrefix := "data: "
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("%s\n", line)
			if line == "" {
				continue
			}

			var data string
			if !strings.HasPrefix(line, dataPrefix) {
				data = line
			} else {
				// 错误  {"error_code":6,"error_msg":"No permission to access data"}
				data = strings.TrimPrefix(line, dataPrefix)
			}
			streamPayload := &Completion{}
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				log.Fatalf("failed to decode stream payload: %v", err)
			}
			responseChan <- streamPayload
		}
		if err := scanner.Err(); err != nil {
			log.Println("issue scanning response:", err)
		}
	}()
	// Parse response
	response := Completion{}

	lastResponse := &Completion{}
	for streamResponse := range responseChan {
		if streamResponse.ErrorCode != 0 || streamResponse.FunctionCall.Name != "" {
			lastResponse = streamResponse
			break
		}

		response.Result += streamResponse.Result
		if req.StreamingFunc != nil {
			err := req.StreamingFunc(ctx, []byte(streamResponse.Result))
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
		lastResponse = streamResponse
	}
	if lastResponse.ErrorCode != 0 {
		return nil, errors.New("errcode:" + lastResponse.ErrorMsg + ",errmsg：" + lastResponse.ErrorMsg)
	}
	if lastResponse.FunctionCall.Name != "" {
		return lastResponse, nil
	}
	// update
	lastResponse.Result = response.Result
	lastResponse.Usage.CompletionTokens = lastResponse.Usage.TotalTokens - lastResponse.Usage.PromptTokens
	return lastResponse, nil
}
