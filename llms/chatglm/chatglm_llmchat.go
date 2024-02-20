package chatglm

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/chatglm/internal/chatglm_client"
	"github.com/tmc/langchaingo/schema"
	"reflect"
)

type ChatMessage = chatglm_client.ChatMessage

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *chatglm_client.Client
	usage            []chatglm_client.Usage
}

const (
	RoleAssistant = "assistant"
	RoleUser      = "user"
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

func NewChatWithCallback(handler callbacks.Handler, opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		CallbacksHandler: handler,
		client:           c,
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
		req := &chatglm_client.ChatRequest{
			Model:         opts.Model,
			Prompt:        messagesToClientMessages(messageSet),
			StreamingFunc: opts.StreamingFunc,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			//RequestId : opts.RequestId,
			Incremental: true,
			Ref: chatglm_client.Ref{
				Enable:      o.client.EnableSearch,
				SearchQuery: o.client.SearchQuery,
			},
		}

		result, err := o.client.CreateChat(ctx, req)
		if err != nil {
			return nil, err
		}
		if result.Code != 200 || !result.Success || len(result.Data.Choices) == 0 {
			return nil, ErrEmptyResponse
		}
		generationInfo := make(map[string]any, reflect.ValueOf(result.Data.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Data.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Data.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Data.Usage.TotalTokens
		msg := &schema.AIChatMessage{
			Content: result.Data.Choices[0].Content,
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
		o.usage = append(o.usage, result.Data.Usage)
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
	o.usage = make([]chatglm_client.Usage, 0, 1)
}

func (o *Chat) GetUsage() []Usage {
	return o.usage
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()
	embeddings := make([][]float64, 0, 1)
	for _, input := range inputTexts {
		embedding, use, err := o.client.CreateEmbedding(ctx, &chatglm_client.EmbeddingRequest{
			Prompt: input,
		})
		if err != nil {
			return nil, err
		}
		if len(embedding) == 0 {
			return nil, ErrEmptyResponse
		}
		embeddings = append(embeddings, embedding)
		o.usage = append(o.usage, use)

	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
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

func messagesToClientMessages(messages []schema.ChatMessage) []*chatglm_client.ChatMessage {
	msgs := make([]*chatglm_client.ChatMessage, len(messages))
	for i, m := range messages {
		msg := &chatglm_client.ChatMessage{
			Content: m.GetContent(),
		}
		typ := m.GetType()
		switch typ {
		//case schema.ChatMessageTypeSystem:
		//	msg.Role = "system"
		case schema.ChatMessageTypeAI:
			msg.Role = "assistant"
		case schema.ChatMessageTypeHuman:
			msg.Role = "user"
		case schema.ChatMessageTypeGeneric:
			msg.Role = "user"
			//case schema.ChatMessageTypeFunction:
			//	msg.Role = "function"
		}
		msgs[i] = msg
	}

	return msgs
}
