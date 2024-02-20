package minimaxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type EmbeddingPayload struct {
	Model string   `json:"model"`
	Texts []string `json:"texts"`
	Type  string   `json:"type"` // db: 存储，query：检索
}

type EmbeddingResponsePayload struct {
	Vectors     [][]float64 `json:"vectors"` // 一个文本对应一个float32数组，长度为1536
	TotalTokens int64       `json:"total_tokens"`
	BaseResp    BaseResp    `json:"base_resp"`
}

// nolint:lll
func (c *Client) CreateEmbedding(ctx context.Context, payload *EmbeddingPayload) (*EmbeddingResponsePayload, error) {
	if payload.Model == "" {
		payload.Model = c.embeddingsModel
	}
	if payload.Type == "" {
		return nil, errors.New("type 参数不能为空，db/query 二选一")
	}
	fmt.Printf("%#v\n", payload)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	if c.baseUrl == "" {
		c.baseUrl = defaultBaseUrl
	}
	url := fmt.Sprintf("%s/embeddings?GroupId=%s", c.baseUrl, c.groupId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
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

		return nil, errors.New(msg) // nolint:goerr113
	}

	var response EmbeddingResponsePayload

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	fmt.Printf("%#v\n", response)
	return &response, nil
}
