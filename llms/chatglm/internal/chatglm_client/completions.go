package chatglm_client

import "context"

type CompletionRequest struct {
	Model       string  `json:"model,omitempty"`
	Prompt      string  `json:"prompt,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	RequestId   string  `json:"request_id,omitempty"`
	// sse返回需设置streamingFunc
	// 结束时返回一个错误 Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	Ref           Ref                                           `json:"ref,omitempty"`
}
type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
type Completion struct {
	Text  string `json:"text"`
	Usage Usage
}

type CompletionResponse struct {
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Success bool   `json:"success,omitempty"`
	Data    Data   `json:"data,omitempty"`
}
type Data struct {
	Choices    []*Choices `json:"choices,omitempty"`
	RequestId  string     `json:"request_id,omitempty"`
	TaskId     string     `json:"task_id,omitempty"`
	TaskStatus string     `json:"task_status,omitempty"`
	Usage      Usage      `json:"usage,omitempty"`
}
type Choices struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

func (c *Client) setCompletionDefaults(payload *CompletionRequest) {
	switch {
	// Prefer the model specified in the payload.
	case payload.Model != "":

	// If no model is set in the payload, take the one specified in the client.
	case c.model != "":
		payload.Model = c.model
	// Fallback: use the default model
	default:
		payload.Model = defaultChatModel
	}
}

func (c *Client) createCompletion(ctx context.Context, payload *CompletionRequest) (*ChatResponse, error) {
	c.setCompletionDefaults(payload)
	return c.createChat(ctx, &ChatRequest{
		Model: payload.Model,
		Prompt: []*ChatMessage{
			{Role: "user", Content: payload.Prompt},
		},
		Temperature:   payload.Temperature,
		TopP:          payload.TopP,
		RequestId:     payload.RequestId,
		Incremental:   true, // 增量
		StreamingFunc: payload.StreamingFunc,
		Ref:           payload.Ref,
	})
}
