package hunyuan

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/hunyuan/internal/hysdk"
	"github.com/tmc/langchaingo/schema"
	"os"
	"strconv"
	"time"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrCodeResponse  = errors.New("has error code")
)

type Usage = hysdk.ResponseUsage

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *hysdk.TencentHyChat
	model            string
	usage            []hysdk.ResponseUsage
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)

	options := &options{}

	for _, opt := range opts {
		opt(options)
	}
	if options.modelName == "" {
		options.modelName = defaultModelName
	}

	return &LLM{
		client: c,
		model:  options.modelName,
	}, err
}

func newClient(opts ...Option) (*hysdk.TencentHyChat, error) {
	appId, _ := strconv.ParseInt(os.Getenv(huanyuanAppId), 10, 64)

	options := &options{
		appId:     appId,
		secretId:  os.Getenv(huanyuanSecretId),
		secretKey: os.Getenv(huanyuanSecretKey),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.appId == 0 || options.secretId == "" || options.secretKey == "" {
		return nil, fmt.Errorf(`%w
You can pass auth info by use hunyuan.New(hunyuan.WithAppId(""),hunyuan.WithSecretId(""),hunyuan.WithSecretKey("")) ,
or
export HUANYUAN_APP_ID={appId} 
export HUANYUAN_SECRET_ID={Secret id}
export HUANYUAN_SECRET_KEY={Secret Key}
doc: https://cloud.tencent.com/document/product/1729/97731`, "授权失败")
	}
	credential := hysdk.NewCredential(options.secretId, options.secretKey)
	return hysdk.NewTencentHyChat(options.appId, credential), nil
}

// GeneratePrompt implements llms.LanguageModel.
func (l *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue,
	options ...llms.CallOption,
) (llms.LLMResult, error) {
	return llms.GeneratePrompt(ctx, l, promptValues, options...)
}

// GetNumTokens implements llms.LanguageModel.
func (l *LLM) GetNumTokens(_ string) int {
	// todo: not provided yet
	// see: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Nlks5zkzu
	return -1
}

func (l *LLM) ResetUsage() {
	l.usage = make([]hysdk.ResponseUsage, 0, 1)
}

func (l *LLM) GetUsage() []Usage {
	return l.usage
}

// Call implements llms.LLM.
func (l *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := l.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}

	if len(r) == 0 {
		return "", ErrEmptyResponse
	}

	return r[0].Text, nil
}

// Generate implements llms.LLM.
func (l *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	l.ResetUsage()
	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMStart(ctx, prompts)
	}
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(prompts))
	for _, prompt := range prompts {
		queryID := uuid.NewString()
		stream := 1
		if opts.StreamingFunc == nil {
			stream = 0
		}
		request := hysdk.Request{
			Timestamp:     int(time.Now().Unix()),
			Expired:       int(time.Now().Unix()) + 24*60*60,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			Messages:      []hysdk.Message{{Role: "user", Content: prompt}},
			QueryID:       queryID,
			Stream:        stream,
			StreamingFunc: opts.StreamingFunc,
		}
		res, err := l.client.Chat(ctx, request)
		if err != nil {
			return nil, err
		}
		if res.Error.Code != 0 {
			fmt.Printf("tencent hunyuan chat err:%+v\n", res.Error)
			return nil, err
		}
		//synchronize 同步打印message
		//fmt.Println(res.Choices[0].Messages.Content)
		generations = append(generations, &llms.Generation{
			Text: res.Choices[0].Messages.Content,
			GenerationInfo: map[string]any{"PromptTokens": res.Usage.PromptTokens,
				"CompletionTokens": res.Usage.CompletionTokens,
				"TotalTokens":      res.Usage.TotalTokens},
		})
		l.usage = append(l.usage, res.Usage)
	}

	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

// CreateEmbedding use ernie Embedding-V1.
// 1. texts counts less than 16
// 2. text runes counts less than 384
// doc: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu
func (l *LLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, errors.New("hunyuan 不支持 embedding")
}
