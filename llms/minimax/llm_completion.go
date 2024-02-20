package minimax

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	minimaxclient2 "github.com/tmc/langchaingo/llms/minimax/minimaxclient"
	"github.com/tmc/langchaingo/schema"
	"reflect"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *minimaxclient2.Client
	usage            []minimaxclient2.Usage
	chatError        error // 每次模型调用的错误信息
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

// NewChat returns a new OpenAI chat LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		client: c,
	}, err
}

func NewWithCallback(handler callbacks.Handler, opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		CallbacksHandler: handler,
		client:           c,
	}, err
}

// Call requests a chat response for the given messages.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) { // nolint: lll

	r, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint:lll,cyclop
	o.ResetUsage()
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, prompts)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(prompts))

	for _, prompt := range prompts {
		req := &minimaxclient2.CompletionRequest{
			Model: opts.Model,
			Messages: []*minimaxclient2.Message{
				{
					SenderType: "USER",
					SenderName: defaultSendName,
					Text:       prompt,
				},
			},
			StreamingFunc:    opts.StreamingFunc,
			TokensToGenerate: int64(opts.MaxTokens),
			Temperature:      float32(opts.Temperature),
			TopP:             float32(opts.TopP),
			BotSetting: []minimaxclient2.BotSetting{{
				BotName: defaultBotName,
				Content: defaultBotDescription,
			}},
			ReplyConstraints: minimaxclient2.ReplyConstraints{
				SenderType: defaultSendType,
				SenderName: defaultBotName,
			},
			//RequestId : opts.RequestId,
			Stream:            opts.StreamingFunc != nil,
			MaskSensitiveInfo: false, // 对输出中易涉及隐私问题的文本信息进行打码，目前包括但不限于邮箱、域名、链接、证件号、家庭住址等，默认true，即开启打码
		}

		result, err := o.client.CreateCompletion(ctx, req)
		if err != nil {
			return nil, err
		}
		if result.InputSensitive {
			o.SetError(fmt.Sprintf("输入命中敏感词：%s", SensitiveTypeToValue(result.InputSensitiveType)))
			return nil, o.GetError()
		}
		if result.OutputSensitive {
			o.SetError(fmt.Sprintf("输出命中敏感词：%s", SensitiveTypeToValue(result.OutputSensitiveType)))
			return nil, o.GetError()
		}
		if result.BaseResp.StatusCode != 0 || len(result.Choices) == 0 {
			return nil, ErrEmptyResponse
		}
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.TotalTokens
		msg := &schema.AIChatMessage{
			Content: result.Choices[0].Messages[0].Text,
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
		o.usage = append(o.usage, result.Usage)
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

func (o *LLM) GetNumTokens(text string) int {
	return 0
}

func (o *LLM) ResetUsage() {
	o.usage = make([]minimaxclient2.Usage, 0, 1)
}

func (o *LLM) SetError(text string) {
	o.chatError = errors.New(text)
}

func (o *LLM) GetError() error {
	return o.chatError
}

func (o *LLM) GetUsage() []minimaxclient2.Usage {
	return o.usage
}

func (o *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, promptValues, options...)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	return nil, errors.New("select embedding type")
	//o.ResetUsage()
	//
	//result, err := o.client.CreateEmbedding(ctx, &minimaxclient.EmbeddingPayload{
	//	Texts: inputTexts,
	//	Type:  "db", //db query
	//})
	//if err != nil {
	//	return nil, err
	//}
	//return result.Vectors, nil
}

// CreateEmbedding 存储文档向量化
func (o *LLM) CreateDbEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()

	result, err := o.client.CreateEmbedding(ctx, &minimaxclient2.EmbeddingPayload{
		Texts: inputTexts,
		Type:  "db", //db query
	})
	if err != nil {
		return nil, err
	}
	o.usage = append(o.usage, minimaxclient2.Usage{TotalTokens: result.TotalTokens})
	return result.Vectors, nil
}

// CreateQueryEmbedding 查询语句向量化
func (o *LLM) CreateQueryEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()

	result, err := o.client.CreateEmbedding(ctx, &minimaxclient2.EmbeddingPayload{
		Texts: inputTexts,
		Type:  "query", //db query
	})
	if err != nil {
		return nil, err
	}
	o.usage = append(o.usage, minimaxclient2.Usage{TotalTokens: result.TotalTokens})
	return result.Vectors, nil
}
