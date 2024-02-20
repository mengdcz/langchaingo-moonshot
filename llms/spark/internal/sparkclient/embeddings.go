package sparkclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	defaultEmbeddingModel = "spark_embedding"
)

type EmbeddingPayloadUser struct {
	Prompt string `json:"prompt"`
}

type EmbeddingPayload struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Payload struct {
		Text string `json:"text"`
	}
}

type EmbeddingResponsePayload struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
	} `json:"header"`
	Payload struct {
		Vector []float64 `json:"vector"`
	} `json:"payload"`
}

func (c *Client) getParamEmbedding(p *EmbeddingPayloadUser) *EmbeddingPayload {
	c.model = defaultEmbeddingModel
	c.baseUrl = defaultBaseUrl3

	resp := &EmbeddingPayload{}
	resp.Header.AppId = c.appId
	resp.Payload.Text = p.Prompt
	return resp
}

func (c *Client) createEmbedding(ctx context.Context, payloadUser *EmbeddingPayloadUser) (*EmbeddingResponsePayload, error) {
	payload := c.getParamEmbedding(payloadUser)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	apiUrl := assembleAuthUrl(c.baseUrl, c.apiKey, c.appSecret)
	fmt.Println(apiUrl)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
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

	var response EmbeddingResponsePayload

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	fmt.Printf("%#v\n", response)
	return &response, nil
}
