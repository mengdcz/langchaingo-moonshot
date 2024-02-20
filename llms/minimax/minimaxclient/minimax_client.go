package minimaxclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Client struct {
	groupId         string
	apiKey          string
	baseUrl         string
	model           string
	httpClient      Doer
	embeddingsModel string
}

func NewClient(opts ...Option) (*Client, error) {
	c := &Client{}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.groupId == "" {
		return nil, errors.New("group id 不能为空")
	}
	if c.apiKey == "" {
		return nil, errors.New("api key 不能为空")
	}

	if c.baseUrl == "" {
		c.baseUrl = defaultBaseUrl
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c, nil
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {

	if r.Model == "" {
		r.Model = c.model
	}
	url := fmt.Sprintf("%s/text/chatcompletion_pro?GroupId=%s", c.baseUrl, c.groupId)
	fmt.Println("url", url)
	body, e := json.Marshal(r)
	if e != nil {
		return nil, e
	}
	fmt.Printf("%s\n", string(body))
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}

	c.setHeader(req)
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
	buf, err := io.ReadAll(resp.Body)
	fmt.Println(string(buf))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// 设置权限
func (c *Client) setHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}

func parseStreamingCompletionResponse(ctx context.Context, resp *http.Response, request *CompletionRequest) (*Completion, error) {
	scanner := bufio.NewScanner(resp.Body)
	dataPrefix := "data: "
	i := 0
	streamPayload := Completion{}
	for scanner.Scan() {
		i++
		line := scanner.Text()
		fmt.Printf("%d : %s\n", i, line)
		if line == "\n" || line == "" {
			continue
		}
		var data string
		if !strings.HasPrefix(line, dataPrefix) {
			continue
		} else {
			// 错误  {"error_code":6,"error_msg":"No permission to access data"}
			data = strings.TrimPrefix(line, dataPrefix)
		}
		//fmt.Println("data=====", data)
		err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
		if err != nil {
			log.Fatalf("failed to decode stream payload: %v", err)
		}
		if streamPayload.OutputSensitive {
			return nil, errors.New("内容违规")
		}

		if request.StreamingFunc != nil && len(streamPayload.Choices) > 0 && len(streamPayload.Choices[0].Messages) > 0 {
			if streamPayload.Choices[0].FinishReason == "stop" {
				break
			}
			err = request.StreamingFunc(ctx, []byte(streamPayload.Choices[0].Messages[0].Text))
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}

	}
	if err := scanner.Err(); err != nil {
		log.Println("issue scanning response:", err)
	}

	fmt.Println("last last :", streamPayload.Choices[0].Messages[0].Text)
	return &streamPayload, nil
}
