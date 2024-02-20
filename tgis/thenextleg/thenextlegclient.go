package thenextleg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultBaseUrl = "https://api.thenextleg.io/v2"
	theNextLegUrl  = "https://api.thenextleg.io"
)

type TheNextLeg struct {
	baseUrl       string // 请求url
	theNextLegUrl string // 某些接口需要需要用到这个接口
	authToken     string
	httpClient    Doer
}

type Option func(*TheNextLeg)

func WithAuthToken(token string) Option {
	return func(leg *TheNextLeg) {
		leg.authToken = token
	}
}

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// New 创建TheNextLeg 客户端实例
func New(opts ...Option) (*TheNextLeg, error) {
	c := &TheNextLeg{}
	for _, v := range opts {
		v(c)
	}
	if c.baseUrl == "" {
		c.baseUrl = defaultBaseUrl
	}
	if c.theNextLegUrl == "" {
		c.theNextLegUrl = theNextLegUrl
	}
	if c.authToken == "" {
		return nil, errors.New("缺少token")
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c, nil
}

type ImagineRequest struct {
	Msg             string `json:"msg"`
	Ref             string `json:"ref,omitempty"`
	WebhookOverride string `json:"webhookOverride,omitempty"`
	IgnorePrefilter string `json:"ignorePrefilter,omitempty"`
}
type MsgIdResponse struct {
	Success   bool   `json:"success"`
	Msg       string `json:"msg,omitempty"`
	MessageId string `json:"messageId,omitempty"`
	CreateAt  string `json:"createAt,omitempty"`
}

// Imagine 生成图片
func (t *TheNextLeg) Imagine(ctx context.Context, payload *ImagineRequest) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/imagine", t.baseUrl)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodPost, payloadBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MessageResponse 生成图片的response
type MessageResponse struct {
	Progress         any      `json:"progress"` // success  100 , in progress : 37, incomplete : incomplete
	Response         Response `json:"response,omitempty"`
	ProgressImageUrl string   `json:"progressImageUrl,omitempty"` // in progress

}
type Response struct {
	CreatedAt            string   `json:"createdAt,omitempty"`
	Buttons              []string `json:"buttons,omitempty"`
	ImageUrl             string   `json:"imageUrl,omitempty"`
	ImageUrls            []string `json:"imageUrls,omitempty"`
	ButtonMessageId      string   `json:"buttonMessageId,omitempty"`
	OriginatingMessageId string   `json:"originatingMessageId,omitempty"`
	Content              any      `json:"content,omitempty"`
	Type                 string   `json:"type,omitempty"`
	Ref                  string   `json:"ref,omitempty"`
	ResponseAt           string   `json:"responseAt,omitempty"`
	Description          string   `json:"description,omitempty"`
}

// Message 获取任务进入，progress 值100：success， 37：生成进度， "incomplete"：失败，未完成
func (t *TheNextLeg) Message(ctx context.Context, msgId string) (*MessageResponse, error) {
	url := fmt.Sprintf("%s/message/%s", t.baseUrl, msgId)

	var resp MessageResponse
	if _, err := t.doHttp(ctx, url, http.MethodGet, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MessageButton 获取任务进入，progress 值100：success， 37：生成进度， "incomplete"：失败，未完成
func (t *TheNextLeg) MessageButton(ctx context.Context, msgId string) (*MessageResponse, error) {
	url := fmt.Sprintf("%s/message/%s", t.baseUrl, msgId)

	var resp MessageResponse
	if _, err := t.doHttp(ctx, url, http.MethodGet, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type ButtonRequest struct {
	ButtonMessageId string `json:"buttonMessageId"`
	Button          string `json:"button,omitempty"`
	Ref             string `json:"ref,omitempty"`
	WebhookOverride string `json:"webhookOverride,omitempty"`
	Prompt          string `json:"prompt,omitempty"`
	Zoom            string `json:"zoom,omitempty"`
	Ar              string `json:"ar,omitempty"`
}

// Button 根据返回的额button按钮继续操作
func (t *TheNextLeg) Button(ctx context.Context, payload *ButtonRequest) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/button", t.baseUrl)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodPost, payloadBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type FaceSwapRequest struct {
	SourceImg string `json:"sourceImg"`
	TargetImg string `json:"targetImg"`
}

// FaceSwap 换脸 您可以执行从一张图像到另一张图像的面部交换。两张图像必须包含一张人脸，否则只会将最左边的人脸作为输入
func (t *TheNextLeg) FaceSwap(ctx context.Context, payload *FaceSwapRequest) ([]byte, error) {
	url := fmt.Sprintf("%s/face-swap", t.theNextLegUrl)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	//var resp MsgIdResponse
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, payloadBytes, nil); err != nil {
		return nil, err
	}
	return b, err
}

// img2img 和生成图像是一个接口

// ImageBuffer Can be one of U1, U2, U3, or U4. get请求
func (t *TheNextLeg) ImageBuffer(ctx context.Context, buttonMessageId, buttonId string) string {
	return fmt.Sprintf("%s/upscale-img-url?buttonMessageId=%s&button=%s", t.theNextLegUrl, buttonMessageId, buttonId)
}

type BlendRequest struct {
	Urls            []string `json:"urls"`
	Ref             string   `json:"ref,omitempty"`
	WebhookOverride string   `json:"webhookOverride,omitempty"`
}

// Blend 混合
func (t *TheNextLeg) Blend(ctx context.Context, payload *BlendRequest) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/blend", t.baseUrl)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodPost, payloadBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SeedRetrieval   ✉️  实际上是调用button接口，传入“✉️”

type DescribeRequest struct {
	Url             string `json:"url"`
	Ref             string `json:"ref,omitempty"`
	WebhookOverride string `json:"webhookOverride,omitempty"`
}

// Describe  您可以使用 Midjourney 来描述您上传和定义的图像
func (t *TheNextLeg) Describe(ctx context.Context, payload *DescribeRequest) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/describe", t.baseUrl)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodPost, payloadBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type SlashCommandsRequest struct {
	Cmd             string `json:"cmd"`
	Ref             string `json:"ref,omitempty"`
	WebhookOverride string `json:"webhookOverride,omitempty"`
}

// Slash Commands 使用”/“命令发送给midjourney
func (t *TheNextLeg) SlashCommands(ctx context.Context, payload *SlashCommandsRequest) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/slash-commands", t.baseUrl)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodPost, payloadBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Info 获取账户信息，例如还剩余多少fast time
func (t *TheNextLeg) Info(ctx context.Context) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/info", t.baseUrl)

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodGet, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AccountSettings 获取账号设置
func (t *TheNextLeg) AccountSettings(ctx context.Context) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/settings", t.baseUrl)

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodGet, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type SettingsRequest struct {
	SettingsToggle  string `json:"settingsToggle"`
	Ref             string `json:"ref,omitempty"`
	WebhookOverride string `json:"webhookOverride,omitempty"`
}

// SetAccountSettings 获取账号设置
func (t *TheNextLeg) SetAccountSettings(ctx context.Context, payload *SettingsRequest) (*MsgIdResponse, error) {
	url := fmt.Sprintf("%s/settings", t.baseUrl)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp MsgIdResponse
	if _, err := t.doHttp(ctx, url, http.MethodPost, payloadBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *TheNextLeg) doHttp(ctx context.Context, url, method string, payloadBytes []byte, resp any) (b []byte, err error) {
	// Build request
	var body io.Reader
	if payloadBytes != nil {
		body = bytes.NewReader(payloadBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return b, err
	}
	t.setHeader(req)
	r, err := t.httpClient.Do(req)
	if err != nil {
		return b, err
	}
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			fmt.Println(err1)
		}
	}(r.Body)
	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)
		return b, errors.New(msg)
	}
	b, err = io.ReadAll(r.Body)
	if err != nil {
		return b, err
	}
	fmt.Println(string(b))
	if resp == nil {
		return b, err
	}
	if err := json.Unmarshal(b, resp); err != nil {
		return b, err
	}
	return b, err
}

func (t *TheNextLeg) setHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+t.authToken)
	req.Header.Set("Content-Type", "application/json")
}
