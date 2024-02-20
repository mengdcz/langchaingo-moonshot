package ernie

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie/internal/ernieclient"
	"github.com/tmc/langchaingo/schema"
)

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *ernieclient.Client
	usage            []ernieclient.Usage
	model            ModelName
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	options := &options{}

	for _, opt := range opts {
		opt(options)
	}
	return &Chat{
		client: c,
		model:  options.modelName,
	}, err
}

// NewChat returns a new OpenAI chat LLM.
func NewChatWithCallback(handler callbacks.Handler, opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	options := &options{}

	for _, opt := range opts {
		opt(options)
	}
	return &Chat{
		client:           c,
		CallbacksHandler: handler,
		model:            options.modelName,
	}, err
}

// Call requests a chat response for the given messages.
func (o *Chat) Call(ctx context.Context, messages []schema.ChatMessage, options ...llms.CallOption) (*schema.AIChatMessage, error) { // nolint: lll
	r, err := o.Generate(ctx, [][]schema.ChatMessage{messages}, options...)
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, ErrEmptyResponse
	}
	return r[0].Message, nil
}

//nolint:funlen
func (o *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint:lll,cyclop
	o.ResetUsage()
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, getPromptsFromMessageSets(messageSets))
	}
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	fmt.Println(opts.Model)
	fmt.Println(o.getModelPath(opts))
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		msgs, system := messagesToClientMessages(messageSet)
		// 如果存在function， 需要function转换
		result, err := o.client.CreateCompletion(ctx, o.getModelPath(opts), &ernieclient.CompletionRequest{
			Messages:      msgs,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			PenaltyScore:  opts.RepetitionPenalty,
			System:        system,
			Functions:     opts.Functions,
			StreamingFunc: opts.StreamingFunc,
			Stream:        opts.StreamingFunc != nil,
		})
		if err != nil {
			return nil, err
		}
		generations = append(generations, &llms.Generation{
			Text:    result.Result,
			Message: &schema.AIChatMessage{Content: result.Result, FunctionCall: &result.FunctionCall},
			GenerationInfo: map[string]any{"PromptTokens": result.Usage.PromptTokens,
				"CompletionTokens": result.Usage.CompletionTokens,
				"TotalTokens":      result.Usage.TotalTokens},
		})
		o.usage = append(o.usage, result.Usage)
	}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}
	return generations, nil
}

func (o *Chat) GetNumTokens(text string) int {
	return 0
}

func (o *Chat) ResetUsage() {
	o.usage = make([]ernieclient.Usage, 0, 1)
}

func (o *Chat) GetUsage() []Usage {
	return o.usage
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, texts []string) ([][]float64, error) {
	o.ResetUsage()
	resp, e := o.client.CreateEmbedding(ctx, texts)
	if e != nil {
		return nil, e
	}

	if resp.ErrorCode > 0 {
		return nil, fmt.Errorf("%w, error_code:%v, erro_msg:%v",
			ErrCodeResponse, resp.ErrorCode, resp.ErrorMsg)
	}

	emb := make([][]float64, 0, len(texts))
	for i := range resp.Data {
		emb = append(emb, resp.Data[i].Embedding)
		o.usage = append(o.usage, resp.Usage)
	}

	return emb, nil
}

func getPromptsFromMessageSets(messageSets [][]schema.ChatMessage) []string {
	prompts := make([]string, 0, len(messageSets))
	for i := 0; i < len(messageSets); i++ {
		curPrompt := ""
		for j := 0; j < len(messageSets[i]); j++ {
			curPrompt += messageSets[i][j].GetContent()
		}
		prompts = append(prompts, curPrompt)
	}

	return prompts
}

func messagesToClientMessages(messages []schema.ChatMessage) ([]*ernieclient.Message, string) {
	msgs := make([]*ernieclient.Message, 0, 4)
	var system string

	for _, m := range messages {
		msg := &ernieclient.Message{}
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem:
			msg.Role = "system"
		case schema.ChatMessageTypeAI:
			msg.Role = "assistant"
		case schema.ChatMessageTypeHuman:
			msg.Role = "user"
		case schema.ChatMessageTypeGeneric:
			msg.Role = "user"
		case schema.ChatMessageTypeFunction:
			msg.Role = "function"
		}
		msg.Content = m.GetContent()
		if n, ok := m.(schema.Named); ok {
			msg.Name = n.GetName()
		}
		// 断言 AIChatMessage
		if mm, ok := m.(schema.AIChatMessage); ok {
			msg.FunctionCall = mm.FunctionCall
		}
		if msg.Role == "system" {

			system = msg.Content
			continue
		}
		msgs = append(msgs, msg)
	}

	fmt.Println(msgs)
	return msgs, system
}

func (o *Chat) getModelPath(opts llms.CallOptions) ernieclient.ModelPath {
	var model ModelName
	model = ModelName(opts.Model)
	if model == "" {
		model = o.model
	}
	if model == "" {
		model = ernieclient.DefaultCompletionModelPath
	}

	switch model {
	case ModelNameERNIEBot4:
		return "completions_pro"
	case ModelNameERNIEBot:
		return "completions"
	case ModelNameERNIEBotTurbo:
		return "eb-instant"
	case ModelNameBloomz7B:
		return "bloomz_7b1"
	case ModelNameLlama2_7BChat:
		return "llama_2_7b"
	case ModelNameLlama2_13BChat:
		return "llama_2_13b"
	case ModelNameLlama2_70BChat:
		return "llama_2_70b"
	default:
		return ernieclient.DefaultCompletionModelPath
	}
}
