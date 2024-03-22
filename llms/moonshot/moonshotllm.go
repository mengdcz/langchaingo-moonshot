package moonshot

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/moonshot/internal/moonshotclient"
	"github.com/tmc/langchaingo/schema"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *moonshotclient.Client
	usage            []moonshotclient.ChatUsage
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

type Usage = moonshotclient.ChatUsage

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
	//if o.CallbacksHandler != nil {
	//	o.CallbacksHandler.HandleLLMStart(ctx, prompts)
	//}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(prompts))
	for _, prompt := range prompts {
		result, err := o.client.CreateCompletion(ctx, &moonshotclient.CompletionRequest{
			Model:         opts.Model,
			Prompt:        prompt,
			MaxTokens:     opts.MaxTokens,
			Temperature:   opts.Temperature,
			N:             opts.N,
			TopP:          opts.TopP,
			StreamingFunc: opts.StreamingFunc,
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
	o.usage = make([]moonshotclient.ChatUsage, 0, 1)
}

func (o *LLM) GetUsage() []Usage {
	return o.usage
}
