package qwenclient

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
	"strconv"
	"strings"
)

const (
	defaultChatModel   = "qwen-turbo" // qwen-turbo   qwen-plus
	FinishReasonNull   = "null"       //null：生成过程中
	FinishReasonStop   = "stop"       //stop token导致结束
	FinishReasonLength = "length"     //生成长度导致结束
)

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type ChatRequestUser struct {
	Model         string                                        `json:"model"`
	BaseURL       string                                        `json:"base_url"`
	Messages      []*ChatMessage                                `json:"messages"`
	ResultFormat  string                                        `json:"result_format,omitempty"` //text"表示旧版本的text "message"表示兼容openai的message
	TopP          float64                                       `json:"top_p,omitempty"`         // (0,1.0)
	TopK          int                                           `json:"top_k,omitempty"`         // (0, 100)
	Send          uint64                                        `json:"send,omitempty"`          // 默认1234
	Temperature   float64                                       `json:"temperature,omitempty"`   // (0,2) 默认1.0
	EnableSearch  bool                                          `json:"enable_search,omitempty"` //生成时，是否参考夸克搜索的结果。注意：打开搜索并不意味着一定会使用搜索结果；如果打开搜索，模型会将搜索结果作为prompt，进而“自行判断”是否生成结合搜索结果的文本，默认为false
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type ChatRequest struct {
	Model         string                                        `json:"model"`
	Input         Input                                         `json:"input"`
	Parameters    Parameters                                    `json:"parameters,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type Input struct {
	Messages []*ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Parameters struct {
	ResultFormat string  `json:"result_format,omitempty"`
	TopP         float64 `json:"top_p,omitempty"`         // (0,1.0)
	TopK         int     `json:"top_k,omitempty"`         // (0, 100)
	Send         uint64  `json:"send,omitempty"`          // 默认1234
	Temperature  float64 `json:"temperature,omitempty"`   // (0,2) 默认1.0
	EnableSearch bool    `json:"enable_search,omitempty"` //生成时，是否参考夸克搜索的结果。注意：打开搜索并不意味着一定会使用搜索结果；如果打开搜索，模型会将搜索结果作为prompt，进而“自行判断”是否生成结合搜索结果的文本，默认为false
}

type ChatResponse struct {
	Output    Output `json:"output"`
	Usage     Usage  `json:"usage"`
	RequestId string `json:"request_id,omitempty"`
}

type Output struct {
	FinishReason string     `json:"finish_reason"`
	Choices      []*Choices `json:"choices,omitempty"`
}
type Usage struct {
	OutputTokens int `json:"output_tokens"`
	InputTokens  int `json:"input_tokens"`
}
type Choices struct {
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamedChatResponsePayload struct {
	Id    int      `json:"id"`
	Event string   `json:"event"`
	Data  *SseData `json:"data"`
}
type SseData struct {
	Output struct {
		Choices []*Choices `json:"choices"`
	}
	Usage     Usage
	RequestId string `json:"request_id"`
	Message   string `json:"message,omitempty"`
	Code      string `json:"code,omitempty"`
}

func (c *Client) getParam(p *ChatRequestUser) *ChatRequest {
	if p.Model == "" {
		p.Model = defaultChatModel
	}
	if p.BaseURL == "" {
		p.BaseURL = defaultBaseUrl
		c.baseURL = p.BaseURL
	}
	if p.ResultFormat == "" {
		p.ResultFormat = "message"
	}
	return &ChatRequest{
		Model: p.Model,
		Input: Input{
			Messages: p.Messages,
		},
		Parameters: Parameters{
			ResultFormat: p.ResultFormat,
			TopP:         p.TopP,
			TopK:         p.TopK,
			Temperature:  p.Temperature,
			Send:         p.Send,
			EnableSearch: p.EnableSearch,
		},
		StreamingFunc: p.StreamingFunc,
	}
}

func (c *Client) createChat(ctx context.Context, payloadUser *ChatRequestUser) (*ChatResponse, error) {
	payload := c.getParam(payloadUser)
	sseEnable := false
	if payload.StreamingFunc != nil {
		sseEnable = true
	}
	payloadBytes, err := json.Marshal(payload)
	fmt.Println(string(payloadBytes))
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, body)
	if err != nil {
		return nil, err
	}
	c.setHeader(req, sseEnable)
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
	aa, _ := io.ReadAll(r.Body)
	var response ChatResponse
	err2 := json.Unmarshal(aa, &response)
	if err2 != nil {
		fmt.Println("json.Unmarshal(aa, &response)")
		fmt.Println(err2.Error())
	}
	return &response, err2

}

func (c *Client) setHeader(req *http.Request, sseEnable bool) {
	req.Header.Set("Content-Type", "application/json")
	if sseEnable {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	} else {
		req.Header.Set("Accept", "*/*")
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
			//log.Println(line)
			//fmt.Println(line)
			//fmt.Println("line====aa")
			//fmt.Println(line)
			//fmt.Println("line====bb")
			// 空行是消息单元结束标志
			if line == "" {
				if unitMsg.Event == "error" {
					log.Println("模型调用报错：", unitMsg.Data.Message)
					break
				}
				//fmt.Printf("%#v\n", unitMsg.Data.Output.Choices[0])
				responseChan <- unitMsg
				// 发送一个消息单元后重新初始化消息单元
				unitMsg = StreamedChatResponsePayload{}
				continue
			}
			err := decodeStreamData(line, &unitMsg)
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
		Output: Output{
			Choices: []*Choices{
				{},
			},
		},
	}

	preLen := 0
	for streamResponse := range responseChan {
		var content string
		//fmt.Println(streamResponse.Data)
		if len(streamResponse.Data.Output.Choices) == 0 {
			continue
		}
		conRune := []rune(streamResponse.Data.Output.Choices[0].Message.Content)
		conRuneLen := len(conRune)

		//content = (streamResponse.Data.Output.Choices[0].Message.Content)[preLen:]

		if streamResponse.Data.Output.Choices[0].FinishReason == FinishReasonNull {
			if conRuneLen-2 > preLen {
				content = string(conRune[preLen : conRuneLen-2])
				preLen = conRuneLen - 2
			} else {
				continue
			}

		} else if streamResponse.Data.Output.Choices[0].FinishReason == FinishReasonStop || streamResponse.Data.Output.Choices[0].FinishReason == FinishReasonLength {
			content = string(conRune[preLen:])
			preLen = conRuneLen
			response.Usage = streamResponse.Data.Usage
		}

		//fmt.Println(content)
		chunk := []byte(content)
		response.Output.Choices[0].Message.Content += content

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
func decodeStreamData(line string, resp *StreamedChatResponsePayload) error {
	var event, id, data string

	if strings.HasPrefix(line, "event:") {
		event = strings.TrimPrefix(line, "event:")
		//log.Println(line, "event:", event)
	} else if strings.HasPrefix(line, "id:") {
		id = strings.TrimPrefix(line, "id:")
		//log.Println(line, "id:", id)
	} else if strings.HasPrefix(line, "data:") {
		data = strings.TrimPrefix(line, "data:")
		//log.Println(line, "data:", data)
	}
	if event != "" {
		resp.Event = event
	}
	if id != "" {
		intId, err := strconv.Atoi(id)
		if err == nil {
			resp.Id = intId
		}
	}
	if data != "" {
		//fmt.Println(data)
		sseData := &SseData{}
		err := json.Unmarshal([]byte(data), sseData)
		if err != nil {
			fmt.Println("sseData fail")
			fmt.Println(err.Error())
			return err
		}
		resp.Data = sseData
	}

	return nil
}
