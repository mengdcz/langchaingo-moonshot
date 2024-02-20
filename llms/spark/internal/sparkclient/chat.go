package sparkclient

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defalutChatModel = "spark1.5"
	status0          = 0
	Status1          = 1
	Status2          = 2
)

type ChatRequestUser struct {
	Model         string                                        `json:"model"`
	Uid           string                                        `json:"uid,omitempty"`
	Temperature   float64                                       `json:"temperature,omitempty"` // [0,1]
	TopK          int                                           `json:"top_k,omitempty"`       //取值为[1，6],默认为4
	MaxTokens     int                                           `json:"max_tokens,omitempty"`
	ChatId        string                                        `json:"chat_id,o,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	Messages      []*Text                                       `json:"text"`
	Functions     []llms.FunctionDefinition                     `json:"functions,omitempty"`
}

type ChatRequest struct {
	Header        Header                                        `json:"header"`
	Parameter     Parameter                                     `json:"parameter"`
	Payload       Payload                                       `json:"payload"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}
type Header struct {
	AppId string `json:"app_id"`
	Uid   string `json:"uid,omitempty"`
}
type Parameter struct {
	Chat ParameterChat `json:"chat"`
}
type ParameterChat struct {
	Domain      string  `json:"domain"`
	Temperature float64 `json:"temperature,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	ChatId      string  `json:"chat_id,omitempty"`
}
type Payload struct {
	Message   PayloadMessage `json:"message"`
	Functions Functions      `json:"functions,omitempty"`
}
type Functions struct {
	Text []llms.FunctionDefinition `json:"text,omitempty"`
}
type PayloadMessage struct {
	Text []*Text `json:"text"`
}
type Text struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamedChatResponsePayload struct {
	Header  RespHeader  `json:"header"`
	Payload RespPayload `json:"payload"`
}
type RespHeader struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
	Status  int    `json:"status"`
}
type RespPayload struct {
	Choices Choices `json:"choices"`
	Usage   Usage   `json:"usage"`
}
type Choices struct {
	Status int            `json:"status"`
	Seq    int            `json:"seq"`
	Text   []*ChoicesText `json:"text"`
}
type ChoicesText struct {
	Content      string               `json:"content"`
	Role         string               `json:"role"`
	Index        int                  `json:"index"`
	FunctionCall *schema.FunctionCall `json:"function_call,omitempty"`
}
type Usage struct {
	Text struct {
		QuestionTokens   int `json:"question_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}
}

type ChatResponse struct {
	Text         string               `json:"text"`
	FunctionCall *schema.FunctionCall `json:"function_call,omitempty"`
	Usage        Usage                `json:"usage"`
}

// getParam 设置参数
func (c *Client) getParam(p *ChatRequestUser) *ChatRequest {
	var domain string
	if p.Model != "" {
		c.model = p.Model
	}
	if c.model == modelName15 {
		c.baseUrl = defaultBaseUrl1
		domain = "general"
	} else if c.model == modelName20 {
		c.baseUrl = defaultBaseUrl2
		domain = "generalv2"
	} else if c.model == modelName30 {
		c.baseUrl = defaultBaseUrl30
		domain = "generalv3"
	}
	resp := &ChatRequest{
		Header: Header{
			AppId: c.appId,
		},
		Parameter: Parameter{
			Chat: ParameterChat{
				Domain: domain,
			},
		},
		Payload: Payload{
			Message: PayloadMessage{
				Text: p.Messages,
			},
			Functions: Functions{
				Text: p.Functions,
			},
		},
		StreamingFunc: p.StreamingFunc,
	}
	if p.Uid != "" {
		resp.Header.Uid = p.Uid
	}
	if p.Temperature > 0 && p.Temperature <= 1 {
		resp.Parameter.Chat.Temperature = p.Temperature
	}
	if p.TopK > 1 && p.TopK <= 6 {
		resp.Parameter.Chat.TopK = p.TopK
	}
	if p.MaxTokens > 0 && p.MaxTokens <= 8192 {
		resp.Parameter.Chat.MaxTokens = p.MaxTokens
	}
	if p.ChatId != "" {
		resp.Parameter.Chat.ChatId = p.ChatId
	}
	return resp
}

func (c *Client) createChat(ctx context.Context, payloadUser *ChatRequestUser) (*ChatResponse, error) {
	// 接受完成后讯飞会主动断开
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	apiUrl := assembleAuthUrl(c.baseUrl, c.apiKey, c.appSecret)
	fmt.Println(apiUrl)
	//握手并建立websocket 连接
	conn, resp, err := d.Dial(apiUrl, nil)
	if err != nil {
		fmt.Println("connection error")
		return nil, errors.New(readResp(resp) + err.Error())
	} else if resp.StatusCode != 101 {
		fmt.Println("connection 101")
		return nil, errors.New(readResp(resp) + err.Error())
	}

	payload := c.getParam(payloadUser)
	b, _ := json.Marshal(payload)
	fmt.Println(string(b))

	go func() {
		conn.WriteJSON(payload)
	}()

	response := &ChatResponse{}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read message error:", err)
			return nil, err
		}

		fmt.Println(string(msg))
		var data StreamedChatResponsePayload
		err1 := json.Unmarshal(msg, &data)
		if err1 != nil {
			fmt.Println("Error parsing JSON:", err)
			return nil, err
		}
		//fmt.Println(string(msg))
		//解析数据

		if data.Header.Code != 0 {
			fmt.Printf("%#v\n", data.Payload)
			return nil, errors.New("error:" + data.Header.Message)
		}
		//fmt.Println(data.Payload.Choices.Text[0].Content)
		if len(data.Payload.Choices.Text) > 0 {
			// 全部文本组装
			response.Text += data.Payload.Choices.Text[0].Content
			response.FunctionCall = data.Payload.Choices.Text[0].FunctionCall
			chunk := []byte(data.Payload.Choices.Text[0].Content)
			if payload.StreamingFunc != nil {
				payload.StreamingFunc(ctx, chunk)
			}
		}

		// 最后一条
		if data.Payload.Choices.Status == 2 {
			response.Usage = data.Payload.Usage
			conn.Close()
			break
		}
	}

	time.Sleep(1 * time.Second)
	return response, nil
}

// 创建鉴权url  apikey 即 hmac username
func assembleAuthUrl(hosturl, apiKey, apiSecret string) string {
	ul, err := url.Parse(hosturl)
	if err != nil {
		fmt.Println(err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	// fmt.Println(sgin)
	//签名结果
	sha := HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	// fmt.Println(sha)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	callurl := hosturl + "?" + v.Encode()
	return callurl
}

func HmacWithShaTobase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

func readResp(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("code=%d,body=%s", resp.StatusCode, string(b))
}
