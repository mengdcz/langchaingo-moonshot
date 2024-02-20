package qwenclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	defaultEmbeddingModel = "text-embedding-v1" // 1536 25 2048
)

type EmbeddingPayload struct {
	Model      string  `json:"model"`
	Input      EmbText `json:"input"`
	Parameters struct {
		TextType string `json:"text_type,omitempty"` //取值：query 或者 document，默认值为 document  ,存库使用document， 查询query
	} `json:"parameters,omitempty"`
}
type EmbText struct {
	Texts []string `json:"texts"`
}

type embeddingResponsePayload struct {
	Output struct {
		Embeddings []struct {
			TextIndex int       `json:"text_index"`
			Embedding []float64 `json:"embedding"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	}
	RequestId string `json:"request_id"`
}

func (c *Client) setEmbeddingHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)
}

func (c *Client) createEmbedding(ctx context.Context, payload *EmbeddingPayload) (*embeddingResponsePayload, error) {
	if payload.Model == "" {
		c.embeddingsModel = defaultEmbeddingModel
	}
	c.baseURL = defaultEmbeddingURL
	if payload.Parameters.TextType == "" {
		payload.Parameters.TextType = "document"
	}
	payloadBytes, err := json.Marshal(payload)
	//fmt.Println(string(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setEmbeddingHeader(req)

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	//ss, _ := io.ReadAll(r.Body)
	//fmt.Println(string(ss))

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

	//fmt.Printf("%#v\n", response)
	return &response, nil
}
