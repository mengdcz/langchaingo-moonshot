package chatglm_client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	// chatglm_pro  chatglm_std  chatglm_lite  characterglm(超拟人大模型)
	defaultChatModel = "chatglm_std"
	eventAdd         = "add"
	eventFinish      = "finish"
	eventError       = "error"
	eventnterrupted  = "interrupted"
)

type ChatRequest struct {
	Model       string         `json:"model,omitempty"`
	Prompt      []*ChatMessage `json:"prompt,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	RequestId   string         `json:"request_id,omitempty"`
	// SSE接口调用时，用于控制每次返回内容方式是增量还是全量，不提供此参数时默认为增量返回, true 为增量返回 , false 为全量返回
	Incremental bool `json:"incremental"`
	// sse返回需设置streamingFunc
	// 结束时返回一个错误 Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	Ref           Ref                                           `json:"ref,omitempty"`
}
type Ref struct {
	Enable      bool   `json:"enable,omitempty"`
	SearchQuery string `json:"search_query,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamedChatResponsePayload struct {
	ID    string `json:"ID,omitempty"`
	Event string `json:"event,omitempty"`
	Data  string `json:"data,omitempty"`
	Meta  *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Usage Usage `json:"usage"`
}

type ChatResponse struct {
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Success bool   `json:"success,omitempty"`
	Data    Data   `json:"data,omitempty"`
}

func (c *Client) createChat(ctx context.Context, payload *ChatRequest) (*ChatResponse, error) {
	var method string
	if payload.StreamingFunc != nil {
		method = "sse-invoke"
	} else {
		method = "invoke"
	}
	payloadBytes, err := json.Marshal(payload)
	fmt.Println(string(payloadBytes))
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(c.model, method), body)
	if err != nil {
		return nil, err
	}
	c.setHeader(req)
	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
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
	if payload.StreamingFunc != nil {
		return parseStreamingChatResponse(ctx, r, payload)
	}
	var response ChatResponse
	return &response, json.NewDecoder(r.Body).Decode(&response)
}

func parseStreamingChatResponse(ctx context.Context, r *http.Response, payload *ChatRequest) (*ChatResponse, error) { //nolint:cyclop,lll
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan StreamedChatResponsePayload)
	go func() {
		defer close(responseChan)
		// 消息单元
		unitMsg := StreamedChatResponsePayload{}
		for scanner.Scan() {
			line := scanner.Text()
			//fmt.Println(line)
			// 是消息单元结束标志
			if line == "" {
				if strings.HasSuffix(unitMsg.Data, "\n") {
					unitMsg.Data = unitMsg.Data[0:(len(unitMsg.Data) - 1)]
				}
				responseChan <- unitMsg
				// 发送一个消息单元后重新初始化消息单元
				unitMsg = StreamedChatResponsePayload{}
				continue
			}
			err := decodeStreamData(line, &unitMsg)
			//err := parse(line, &unitMsg)
			if err != nil {
				log.Printf("failed to decode stream payload: %v", err)
				break
			}

		}
		if err := scanner.Err(); err != nil {
			log.Println("issue scanning response:", err)
		}
	}()
	// Parse response
	response := ChatResponse{
		Code:    200,
		Success: true,
		Data: Data{
			Choices: []*Choices{
				{},
			},
			Usage: Usage{},
		},
	}

	for streamResponse := range responseChan {
		//fmt.Printf("%#v\n", streamResponse)
		chunk := []byte(streamResponse.Data)
		if streamResponse.Event == eventAdd {
			response.Data.Choices[0].Content += streamResponse.Data
		} else if streamResponse.Event == eventFinish {
			response.Data.Usage = streamResponse.Meta.Usage
		}

		if payload.StreamingFunc != nil {
			err := payload.StreamingFunc(ctx, chunk)
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
	}
	return &response, nil
}

// 数据格式, 流式数据返回格式
// 2023/09/20 19:11:42 event:add
// 2023/09/20 19:11:42 id:7951544019094669224
// 2023/09/20 19:11:42 data:和支持
// 2023/09/20 19:11:42
// 2023/09/20 19:11:42 event:add
// 2023/09/20 19:11:42 id:7951544019094669224
// 2023/09/20 19:11:42 data:。
// 2023/09/20 19:11:42
// 2023/09/20 19:11:42 event:finish
// 2023/09/20 19:11:42 id:7951544019094669224
// 2023/09/20 19:11:42 data:
// 2023/09/20 19:11:42 meta:{"task_status":"SUCCESS","usage":{"completion_tokens":59,"prompt_tokens":0,"total_tokens":59},"task_id":"7951544019094669224","request_id":"7951544019094669224"}
func decodeStreamData(line string, resp *StreamedChatResponsePayload) error {
	var event, id, data, meta string

	if strings.HasPrefix(line, "event:") {
		event = strings.TrimPrefix(line, "event:")
		//log.Println(line, "event:", event)
	} else if strings.HasPrefix(line, "id:") {
		id = strings.TrimPrefix(line, "id:")
		//log.Println(line, "id:", id)
	} else if strings.HasPrefix(line, "data:") {
		data = strings.TrimPrefix(line, "data:")
		data += "\n"
		//if data == "" {
		//	data = "\n"
		//}
	} else if strings.HasPrefix(line, "meta:") {
		meta = strings.TrimPrefix(line, "meta:")
		//log.Println(line, "meta:", meta)
	}
	if event != "" {
		resp.Event = event
	}
	if id != "" {
		resp.ID = id
	}
	resp.Data += data

	if meta != "" {
		metaS := &Meta{}
		err := json.Unmarshal([]byte(meta), metaS)
		if err != nil {
			return err
		}
		resp.Meta = metaS
	}
	return nil
}
