package spark

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/spark/internal/sparkclient"
	"github.com/tmc/langchaingo/schema"
)

type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *sparkclient.Client
	usage            []Usage
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
		result, err := o.client.CreateCompletion(ctx, &sparkclient.ChatRequestUser{
			Model:         opts.Model,
			Messages:      []*sparkclient.Text{&sparkclient.Text{Role: "user", Content: prompt}},
			MaxTokens:     opts.MaxTokens,
			Temperature:   opts.Temperature,
			TopK:          opts.TopK,
			StreamingFunc: opts.StreamingFunc,
			Functions:     opts.Functions,
		})
		if err != nil {
			return nil, err
		}
		generations = append(generations, &llms.Generation{
			Text: result.Text,
			GenerationInfo: map[string]any{
				"PromptTokens":     result.Usage.Text.PromptTokens,
				"CompletionTokens": result.Usage.Text.CompletionTokens,
				"TotalTokens":      result.Usage.Text.TotalTokens},
		})
		o.usage = append(o.usage, Usage{
			PromptTokens:     result.Usage.Text.PromptTokens,
			CompletionTokens: result.Usage.Text.CompletionTokens,
			TotalTokens:      result.Usage.Text.TotalTokens,
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
	o.usage = make([]Usage, 0, 1)
}

func (o *LLM) GetUsage() []Usage {
	return o.usage
}

func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings := make([][]float64, 0, 1)
	o.ResetUsage()
	for _, input := range inputTexts {
		embedding, err := o.client.CreateEmbedding(ctx, &sparkclient.EmbeddingPayloadUser{
			Prompt: input,
		})
		if err != nil {
			return nil, err
		}
		if len(embedding) == 0 {
			return nil, ErrEmptyResponse
		}
		embeddings = append(embeddings, embedding)
		// 用于记录本次token使用情况 , 讯飞embedding没有返回usage
		o.usage = append(o.usage, Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		})

	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
}
