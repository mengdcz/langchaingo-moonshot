package spark

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/spark/internal/sparkclient"
	"github.com/tmc/langchaingo/schema"
	"reflect"
)

type ChatMessage = sparkclient.Text

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *sparkclient.Client
	Usage            []Usage
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

func NewChat(opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		client: c,
	}, err
}
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
		req := &sparkclient.ChatRequestUser{
			Model:         opts.Model,
			Messages:      messagesToClientMessages(messageSet),
			StreamingFunc: opts.StreamingFunc,
			Temperature:   opts.Temperature,
			MaxTokens:     opts.MaxTokens,
			TopK:          opts.TopK,
			Functions:     opts.Functions,
		}
		result, err := o.client.CreateChat(ctx, req)
		if err != nil {
			return nil, err
		}
		//if len(result.Text) == 0 {
		//	return nil, ErrEmptyResponse
		//}
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage.Text).NumField())
		generationInfo["CompletionTokens"] = result.Usage.Text.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.Text.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.Text.TotalTokens
		msg := &schema.AIChatMessage{
			Content:      result.Text,
			FunctionCall: result.FunctionCall,
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
		o.Usage = append(o.Usage, Usage{
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

func (o *Chat) GetNumTokens(text string) int {
	return 0
}

func (o *Chat) ResetUsage() {
	o.Usage = make([]Usage, 0, 1)
}

func (o *Chat) GetUsage() []Usage {
	return o.Usage
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()
	embeddings := make([][]float64, 0, 1)
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
		// 讯飞星火embedding没有返回usage
		o.Usage = append(o.Usage, Usage{})

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

func messagesToClientMessages(messages []schema.ChatMessage) []*ChatMessage {
	msgs := make([]*ChatMessage, len(messages))
	for i, m := range messages {
		msg := &ChatMessage{
			Content: m.GetContent(),
		}
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeAI:
			msg.Role = "assistant"
		case schema.ChatMessageTypeHuman:
			msg.Role = "user"
		}
		msgs[i] = msg
	}

	return msgs
}
