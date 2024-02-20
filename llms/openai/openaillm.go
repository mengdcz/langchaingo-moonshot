package openai

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/schema"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *openaiclient.Client
	usage            []openaiclient.ChatUsage
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

type Usage = openaiclient.ChatUsage

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		client: c,
	}, err
}

// Call requests a completion for the given prompt.
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
		result, err := o.client.CreateCompletion(ctx, &openaiclient.CompletionRequest{
			Model:            opts.Model,
			Prompt:           prompt,
			MaxTokens:        opts.MaxTokens,
			StopWords:        opts.StopWords,
			Temperature:      opts.Temperature,
			N:                opts.N,
			FrequencyPenalty: opts.FrequencyPenalty,
			PresencePenalty:  opts.PresencePenalty,
			TopP:             opts.TopP,
			StreamingFunc:    opts.StreamingFunc,
		})
		if err != nil {
			return nil, err
		}
		generations = append(generations, &llms.Generation{
			Text: result.Text,
		})
		PromptTokens := o.GetNumTokens(prompt)
		CompletionTokens := o.GetNumTokens(result.Text)
		TotalTokens := PromptTokens + CompletionTokens
		o.usage = append(o.usage, Usage{
			PromptTokens:     PromptTokens,
			CompletionTokens: CompletionTokens,
			TotalTokens:      TotalTokens,
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

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens(o.client.Model, text)
}

func (o *LLM) ResetUsage() {
	o.usage = make([]openaiclient.ChatUsage, 0, 1)
}

func (o *LLM) GetUsage() []Usage {
	return o.usage
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
		Input: inputTexts,
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
	total := o.GetNumTokens(strings.Join(inputTexts, ""))
	o.usage = []openaiclient.ChatUsage{
		{
			PromptTokens:     total,
			CompletionTokens: 0,
			TotalTokens:      total,
		},
	}
	return embeddings, nil
}
