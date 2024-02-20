package qwen

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
	"github.com/tmc/langchaingo/schema"
)

type Usage = qwenclient.Usage

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *qwenclient.Client
	usage            []CompletionUsage
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		client: c,
	}, err
}

func NewWithCallback(handler callbacks.Handler, opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		CallbacksHandler: handler,
		client:           c,
	}, err
}

func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	// 每次调用前清空usage
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
		result, err := o.client.CreateCompletion(ctx, &qwenclient.ChatRequestUser{

			Model: opts.Model,
			Messages: []*qwenclient.ChatMessage{
				{Role: "user", Content: prompt},
			},
			Temperature: opts.Temperature,
			TopP:        opts.TopP,
			TopK:        opts.TopK,
			//RequestId:     opts.RequestId,
			StreamingFunc: opts.StreamingFunc,
			EnableSearch:  o.client.EnableSearch,
		})
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		if len(result.Output.Choices) == 0 {
			continue
		}
		generations = append(generations, &llms.Generation{
			Text: result.Output.Choices[0].Message.Content,
			GenerationInfo: map[string]any{"PromptTokens": result.Usage.InputTokens,
				"CompletionTokens": result.Usage.OutputTokens,
				"TotalTokens":      result.Usage.InputTokens + result.Usage.OutputTokens},
		})
		// 将本次调用的usage暂存
		o.usage = append(o.usage, CompletionUsage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		})
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

func (o *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, promptValues, options...)
}

// 与openai的实现方式不同, 无法计算
func (o *LLM) GetNumTokens(text string) int {
	return 0
}

func (o *LLM) ResetUsage() {
	o.usage = make([]CompletionUsage, 0, 1)
}

func (o *LLM) GetUsage() []CompletionUsage {
	return o.usage
}

func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()
	embeddings, use, err := o.client.CreateEmbedding(ctx, &qwenclient.EmbeddingPayload{
		Input: qwenclient.EmbText{
			Texts: inputTexts,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	o.usage = append(o.usage, CompletionUsage{
		TotalTokens: use,
	})
	return embeddings, nil
}
