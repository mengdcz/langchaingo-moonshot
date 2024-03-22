package moonshot

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/moonshot/internal/moonshotclient"
	"github.com/tmc/langchaingo/schema"
	"reflect"
)

type ChatMessage = moonshotclient.ChatMessage

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *moonshotclient.Client
	usage            []moonshotclient.ChatUsage
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
	//if o.CallbacksHandler != nil {
	//	o.CallbacksHandler.HandleLLMStart(ctx, getPromptsFromMessageSets(messageSets))
	//}
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		req := &moonshotclient.ChatRequest{
			Model:         opts.Model,
			Messages:      messagesToClientMessages(messageSet),
			StreamingFunc: opts.StreamingFunc,
			Temperature:   opts.Temperature,
			MaxTokens:     opts.MaxTokens,
			N:             opts.N, // TODO: note, we are not returning multiple completions

			//FunctionCallBehavior: moonshotclient.FunctionCallBehavior(opts.FunctionCallBehavior),
		}

		result, err := o.client.CreateChat(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(result.Choices) == 0 {
			return nil, ErrEmptyResponse
		}
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.TotalTokens
		msg := &schema.AIChatMessage{
			Content: result.Choices[0].Message.Content,
		}

		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
		PromptTokens := result.Usage.PromptTokens         //o.countMessageTokens(messageSet)
		CompletionTokens := result.Usage.CompletionTokens //o.GetNumTokens(msg.Content)
		TotalTokens := result.Usage.TotalTokens           // PromptTokens + CompletionTokens
		o.usage = append(o.usage, Usage{
			PromptTokens:     PromptTokens,
			CompletionTokens: CompletionTokens,
			TotalTokens:      TotalTokens,
		})
	}
	//if o.CallbacksHandler != nil {
	//	o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	//}
	return generations, nil
}

func (o *Chat) GetNumTokens(text string) int {
	return 0 // llms.CountTokens(o.client.Model, text)
}

func (o *Chat) ResetUsage() {
	o.usage = make([]moonshotclient.ChatUsage, 0, 1)
}

func (o *Chat) GetUsage() []Usage {
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

func messagesToClientMessages(messages []schema.ChatMessage) []*moonshotclient.ChatMessage {
	msgs := make([]*moonshotclient.ChatMessage, len(messages))
	for i, m := range messages {
		msg := &moonshotclient.ChatMessage{
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
		if n, ok := m.(schema.Named); ok {
			msg.Name = n.GetName()
		}
		msgs[i] = msg
	}

	return msgs
}

func (o *Chat) countMessageTokens(messages []schema.ChatMessage) int {
	content := ""
	for _, v := range messages {
		content += v.GetContent()
	}
	return o.GetNumTokens(content)
}
