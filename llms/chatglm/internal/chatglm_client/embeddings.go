package chatglm_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	defaultEmbeddingModel = "text_embedding"
)

type embeddingPayload struct {
	Prompt    string `json:"prompt"`
	RequestId string `json:"request_id,omitempty"`
}

type embeddingResponsePayload struct {
	Code    int           `json:"code"`
	Msg     string        `json:"msg"`
	Success bool          `json:"success"`
	Data    EmbeddingData `json:"data"`
}

type EmbeddingData struct {
	Embedding  []float64 `json:"embedding"`
	RequestId  string    `json:"request_id"`
	TaskId     string    `json:"task_id"`
	TaskStatus string    `json:"task_status"`
	Usage      Usage     `json:"usage"`
}

// nolint:lll
func (c *Client) createEmbedding(ctx context.Context, payload *embeddingPayload) (*embeddingResponsePayload, error) {
	fmt.Printf("%#v\n", payload)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(defaultEmbeddingModel, "invoke"), bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeader(req)

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg) // nolint:goerr113
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
	}

	var response embeddingResponsePayload

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	fmt.Printf("%#v\n", response)
	return &response, nil
}
