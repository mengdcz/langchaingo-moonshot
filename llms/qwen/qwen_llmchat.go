package qwen

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
	"github.com/tmc/langchaingo/schema"
	"reflect"
)

type ChatMessage = qwenclient.ChatMessage

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *qwenclient.Client
	usage            []CompletionUsage
}

const (
	RoleSystem    = "system"
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleFunction  = "function"
)

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		client: c,
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
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		req := &qwenclient.ChatRequestUser{
			Model:         opts.Model,
			Messages:      messagesToClientMessages(messageSet),
			StreamingFunc: opts.StreamingFunc,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			TopK:          opts.TopK,
			EnableSearch:  o.client.EnableSearch,
		}
		result, err := o.client.CreateChat(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(result.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
		}
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.OutputTokens
		generationInfo["PromptTokens"] = result.Usage.InputTokens
		generationInfo["TotalTokens"] = result.Usage.OutputTokens + result.Usage.InputTokens
		msg := &schema.AIChatMessage{
			Content: result.Output.Choices[0].Message.Content,
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
		o.usage = append(o.usage, CompletionUsage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.OutputTokens + result.Usage.InputTokens,
		})
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
	o.usage = make([]CompletionUsage, 0, 1)
}

func (o *Chat) GetUsage() []CompletionUsage {
	return o.usage
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
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

func messagesToClientMessages(messages []schema.ChatMessage) []*qwenclient.ChatMessage {
	msgs := make([]*qwenclient.ChatMessage, len(messages))
	for i, m := range messages {
		msg := &qwenclient.ChatMessage{
			Content: m.GetContent(),
		}
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

		msgs[i] = msg
	}

	return msgs
}

func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
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
