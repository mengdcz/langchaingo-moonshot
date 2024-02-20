package hunyuan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/hunyuan/internal/hysdk"
	"github.com/tmc/langchaingo/schema"
	"time"
)

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *hysdk.TencentHyChat
	usage            []hysdk.ResponseUsage
	model            string
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
	if options.modelName == "" {
		options.modelName = defaultModelName
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
	if options.modelName == "" {
		options.modelName = defaultModelName
	}
	if options.appId == 0 || options.secretId == "" || options.secretKey == "" {
		return nil, fmt.Errorf(`%w
You can pass auth info by use hunyuan.New(hunyuan.WithAppId(""),hunyuan.WithSecretId(""),hunyuan.WithSecretKey("")) ,
or
export HUANYUAN_APP_ID={appId} 
export HUANYUAN_SECRET_ID={Secret id}
export HUANYUAN_SECRET_KEY={Secret Key}
doc: https://cloud.tencent.com/document/product/1729/97731`, "授权失败")
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
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		// 混元不支持 system
		msgs, _ := messagesToClientMessages(messageSet)

		queryID := uuid.NewString()
		stream := 1
		if opts.StreamingFunc == nil {
			stream = 0
		}
		request := hysdk.Request{
			Timestamp:     int(time.Now().Unix()),
			Expired:       int(time.Now().Unix()) + 24*60*60,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			Messages:      msgs,
			QueryID:       queryID,
			Stream:        stream,
			StreamingFunc: opts.StreamingFunc,
		}
		res, err := o.client.Chat(ctx, request)
		if err != nil {
			fmt.Println("chat error:", err.Error())
			return nil, err
		}
		if res.Error.Code != 0 {
			fmt.Printf("tencent hunyuan chat err:%+v\n", res.Error)
			return nil, err
		}
		//synchronize 同步打印message
		b, _ := json.Marshal(res)
		fmt.Println("generation:", string(b))
		generations = append(generations, &llms.Generation{
			Text:    res.Choices[0].Delta.Content,
			Message: &schema.AIChatMessage{Content: res.Choices[0].Delta.Content},
			GenerationInfo: map[string]any{"PromptTokens": res.Usage.PromptTokens,
				"CompletionTokens": res.Usage.CompletionTokens,
				"TotalTokens":      res.Usage.TotalTokens},
		})
		o.usage = append(o.usage, res.Usage)
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
	o.usage = make([]hysdk.ResponseUsage, 0, 1)
}

func (o *Chat) GetUsage() []Usage {
	return o.usage
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, texts []string) ([][]float64, error) {

	return nil, errors.New("hunyuan unimpl embedding")
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

func messagesToClientMessages(messages []schema.ChatMessage) ([]hysdk.Message, string) {
	msgs := make([]hysdk.Message, 0, 4)
	var system string

	for _, m := range messages {
		msg := hysdk.Message{}
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
		msg.Content = m.GetContent()
		//if n, ok := m.(schema.Named); ok {
		//	msg.Name = n.GetName()
		//}
		//// 断言 AIChatMessage
		//if mm, ok := m.(schema.AIChatMessage); ok {
		//	msg.FunctionCall = mm.FunctionCall
		//}
		//if msg.Role == "system" {
		//
		//	system = msg.Content
		//	continue
		//}
		msgs = append(msgs, msg)
	}

	fmt.Println(msgs)
	return msgs, system
}
