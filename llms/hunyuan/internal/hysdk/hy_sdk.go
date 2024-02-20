/*
 * Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package hysdk ...
package hysdk

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	protocol = "https"
	host     = "hunyuan.cloud.tencent.com"
	path     = "/hyllm/v1/chat/completions?"
)

const (
	//Synchronize 同步
	Synchronize = iota
	//Stream 流式
	Stream
)

func getUrl() string {
	return fmt.Sprintf("%s://%s%s", protocol, host, path)
}

func getFullPath() string {
	return host + path
}

// ResponseChoices :结果
type ResponseChoices struct {
	FinishReason string  `json:"finish_reason,omitempty"` //流式结束标志位，为 stop 则表示尾包
	Messages     Message `json:"messages,omitempty"`      //内容，同步模式返回内容，流模式为 null 输出 content 内容总数最多支持 1024token。
	Delta        Message `json:"delta,omitempty"`         //内容，流模式返回内容，同步模式为 null 输出 content 内容总数最多支持 1024token。
}

// ResponseUsage :token 数量
type ResponseUsage struct {
	PromptTokens     int64 `json:"prompt_tokens,omitempty"`     //输入 token 数量
	TotalTokens      int64 `json:"total_tokens,omitempty"`      //token 数量
	CompletionTokens int64 `json:"completion_tokens,omitempty"` //输出 token 数量
}

// ResponseError :错误信息
type ResponseError struct {
	Message string `json:"message,omitempty"` //错误提示信息
	Code    int    `json:"code,omitempty"`    //Code 错误码
}

// Message :会话内容 同步调用返回
type Message struct {
	//Role 当前支持以下：
	//user：表示用户
	//assistant：表示对话助手
	//在 message 中必须是 user 与 assistant 交替(一问一答)
	Role    string `json:"role"`
	Content string `json:"content"` //消息的内容
}

// ResponseChoicesDelta :流式调用返回
type ResponseChoicesDelta struct {
	Content string `json:"content"`
}

// Request :请求
type Request struct {
	AppID    int64  `json:"app_id"`    //腾讯云账号的 APPID
	SecretID string `json:"secret_id"` //官网 SecretId
	//Timestamp当前 UNIX 时间戳，单位为秒，可记录发起 API 请求的时间。
	//例如1529223702，如果与当前时间相差过大，会引起签名过期错误
	Timestamp int `json:"timestamp"`
	//Expired 签名的有效期，是一个符合 UNIX Epoch 时间戳规范的数值，
	//单位为秒；Expired 必须大于 Timestamp 且 Expired-Timestamp 小于90天
	Expired int    `json:"expired"`
	QueryID string `json:"query_id"` //请求 ID，用于问题排查
	// Temperature 较高的数值会使输出更加随机，而较低的数值会使其更加集中和确定
	// 默认1.0，取值区间为[0.0, 2.0]，非必要不建议使用, 不合理的取值会影响效果
	// 建议该参数和top_p只设置1个，不要同时更改 top_p
	Temperature float64 `json:"temperature"`
	// TopP 影响输出文本的多样性，取值越大，生成文本的多样性越强
	// 默认1.0，取值区间为[0.0, 1.0]，非必要不建议使用, 不合理的取值会影响效果
	// 建议该参数和 temperature 只设置1个，不要同时更改
	TopP float64 `json:"top_p"`
	// Stream 0：同步，1：流式 （默认，协议：SSE)
	//同步请求超时：60s，如果内容较长建议使用流式
	Stream int `json:"stream"`
	// Messages 会话内容, 长度最多为40, 按对话时间从旧到新在数组中排列
	// 输入 content 总数最大支持 3000 token。
	Messages      []Message                                     `json:"messages"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

// Response :回包
type Response struct {
	Choices []ResponseChoices `json:"choices,omitempty"` // 结果
	Created string            `json:"created,omitempty"` //unix 时间戳的字符串
	ID      string            `json:"id,omitempty"`      //会话 id
	Usage   ResponseUsage     `json:"usage,omitempty"`   //token 数量
	Error   ResponseError     `json:"error,omitempty"`   //错误信息 注意：此字段可能返回 null，表示取不到有效值
	Note    string            `json:"note,omitempty"`    //注释
	ReqID   string            `json:"req_id,omitempty"`  //唯一请求 ID，每次请求都会返回。用于反馈接口入参
}

// Credential :凭据
type Credential struct {
	SecretID  string
	SecretKey string
}

// NewCredential ...
func NewCredential(secretID, secretKey string) *Credential {
	return &Credential{SecretID: secretID, SecretKey: secretKey}
}

// TencentHyChat ...
type TencentHyChat struct {
	Credential *Credential
	AppID      int64
}

// NewTencentHyChat ...
func NewTencentHyChat(appId int64, credential *Credential) *TencentHyChat {
	return &TencentHyChat{
		Credential: credential,
		AppID:      appId,
	}
}

// NewRequest ...
func NewRequest(mod int, messages []Message) Request {
	queryID := uuid.NewString()
	return Request{
		Timestamp:   int(time.Now().Unix()),
		Expired:     int(time.Now().Unix()) + 24*60*60,
		Temperature: 0,
		TopP:        0.8,
		Messages:    messages,
		QueryID:     queryID,
		Stream:      mod,
	}
}

func (t *TencentHyChat) getHttpReq(ctx context.Context, req Request) (*http.Request, error) {
	req.AppID = t.AppID
	req.SecretID = t.Credential.SecretID
	signatureUrl := t.buildURL(req)
	signature := t.genSignature(signatureUrl)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("json marshal err: %+v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", getUrl(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new http request err: %+v", err)
	}
	httpReq.Header.Set("Authorization", signature)
	httpReq.Header.Set("Content-Type", "application/json")

	if req.Stream == Stream {
		httpReq.Header.Set("Cache-Control", "no-cache")
		httpReq.Header.Set("Connection", "keep-alive")
		httpReq.Header.Set("Accept", "text/event-Stream")
	}

	return httpReq, nil
}

// Chat 发起一个 Chat
func (t *TencentHyChat) Chat(ctx context.Context, req Request) (*Response, error) {
	res := make(chan Response, 1)
	httpReq, err := t.getHttpReq(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("do general http request err: %+v", err)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do chat request err: %+v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("do chat request failed status code :%d", httpResp.StatusCode)
	}

	if req.Stream == Synchronize {
		return t.synchronize(ctx, httpResp, res)
	}
	return t.stream(ctx, httpResp, res, req.StreamingFunc)
}

func (t *TencentHyChat) synchronize(ctx context.Context, httpResp *http.Response, res chan Response) (chatResp *Response, err error) {
	defer func() {
		httpResp.Body.Close()
		close(res)
	}()
	respBody, err := io.ReadAll(httpResp.Body)
	fmt.Println("xuanyuan resp body:", string(respBody))
	if err != nil {
		return chatResp, fmt.Errorf("read response body err: %+v", err)
	}
	chatResp = &Response{}
	if err = json.Unmarshal(respBody, chatResp); err != nil {
		return chatResp, fmt.Errorf("json unmarshal err: %+v", err)
	}
	return chatResp, err
}

func (t *TencentHyChat) stream(ctx context.Context, httpResp *http.Response, res chan Response, streamingFunc func(ctx context.Context, chunk []byte) error) (chatResp *Response, err error) {
	defer func() {
		httpResp.Body.Close()
		close(res)
	}()
	go func() {

	}()
	reader := bufio.NewReader(httpResp.Body)
	go func() {
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				res <- Response{Error: ResponseError{Message: fmt.Sprintf("read Stream data err: %+v", err), Code: 500}}
				return
			}

			dataStr := strings.TrimSpace(string(data))
			if dataStr == "" {
				continue
			}

			if !strings.HasPrefix(dataStr, "data: ") {
				continue
			}

			var chatResponse Response
			if err := json.Unmarshal([]byte(dataStr[6:]), &chatResponse); err != nil {
				res <- Response{Error: ResponseError{Message: fmt.Sprintf("json unmarshal err: %+v", err), Code: 500}}
				return
			}

			res <- chatResponse
			if chatResponse.Choices[0].FinishReason == "stop" {
				return
			}
		}
	}()

	// Parse response
	response := Response{
		Choices: []ResponseChoices{{}},
	}

	lastResponse := Response{}
	for streamResponse := range res {
		b, _ := json.Marshal(streamResponse)
		fmt.Println(string(b))

		response.Choices[0].Delta.Content += streamResponse.Choices[0].Delta.Content
		if streamingFunc != nil {
			err := streamingFunc(ctx, []byte(streamResponse.Choices[0].Delta.Content))
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
		lastResponse = streamResponse
		if lastResponse.Choices[0].FinishReason == "stop" {
			break
		}
	}
	if lastResponse.Error.Code != 0 {
		return nil, errors.New("errcode:,errmsg：" + lastResponse.Error.Message)
	}

	// update
	lastResponse.Choices[0].Delta.Content = response.Choices[0].Delta.Content
	//lastResponse.Usage.CompletionTokens = lastResponse.Usage.TotalTokens - lastResponse.Usage.PromptTokens
	return &lastResponse, nil
}

// genSignature 签名
func (t *TencentHyChat) genSignature(url string) string {
	mac := hmac.New(sha1.New, []byte(t.Credential.SecretKey))
	signURL := url
	mac.Write([]byte(signURL))
	sign := mac.Sum([]byte(nil))
	return base64.StdEncoding.EncodeToString(sign)
}

// buildSignURL 构建签名url
func (t *TencentHyChat) buildURL(req Request) string {
	params := make([]string, 0)
	params = append(params, "app_id="+strconv.FormatInt(req.AppID, 10))
	params = append(params, "secret_id="+req.SecretID)
	params = append(params, "timestamp="+strconv.Itoa(req.Timestamp))
	params = append(params, "query_id="+req.QueryID)
	params = append(params, "temperature="+strconv.FormatFloat(req.Temperature, 'f', -1, 64))
	params = append(params, "top_p="+strconv.FormatFloat(req.TopP, 'f', -1, 64))
	params = append(params, "stream="+strconv.Itoa(req.Stream))
	params = append(params, "expired="+strconv.Itoa(req.Expired))

	var messageStr string
	for _, msg := range req.Messages {
		messageStr += fmt.Sprintf(`{"role":"%s","content":"%s"},`, msg.Role, msg.Content)
	}
	messageStr = strings.TrimSuffix(messageStr, ",")
	params = append(params, "messages=["+messageStr+"]")

	sort.Sort(sort.StringSlice(params))
	return getFullPath() + strings.Join(params, "&")
}
