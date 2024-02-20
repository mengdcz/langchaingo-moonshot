package minimax

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	minimaxclient2 "github.com/tmc/langchaingo/llms/minimax/minimaxclient"
	"github.com/tmc/langchaingo/schema"
	"reflect"
)

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *minimaxclient2.Client
	usage            []minimaxclient2.Usage
	chatError        error // 每次模型调用的错误信息
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
		clientMsg, setting, reply := messagesToClientMessages(messageSet)
		req := &minimaxclient2.CompletionRequest{
			Model:            opts.Model,
			Messages:         clientMsg,
			StreamingFunc:    opts.StreamingFunc,
			TokensToGenerate: int64(opts.MaxTokens),
			Temperature:      float32(opts.Temperature),
			TopP:             float32(opts.TopP),
			BotSetting:       []minimaxclient2.BotSetting{setting},
			ReplyConstraints: reply,
			//RequestId : opts.RequestId,
			Stream:            opts.StreamingFunc != nil,
			MaskSensitiveInfo: false, // 对输出中易涉及隐私问题的文本信息进行打码，目前包括但不限于邮箱、域名、链接、证件号、家庭住址等，默认true，即开启打码
			Functions:         opts.Functions,
			//FunctionCallSetting   自动模式等
		}

		result, err := o.client.CreateCompletion(ctx, req)
		if err != nil {
			return nil, err
		}
		if result.InputSensitive {
			o.SetError(fmt.Sprintf("输入命中敏感词：%s", SensitiveTypeToValue(result.InputSensitiveType)))
			return nil, o.GetError()
		}
		if result.OutputSensitive {
			o.SetError(fmt.Sprintf("输出命中敏感词：%s", SensitiveTypeToValue(result.OutputSensitiveType)))
			return nil, o.GetError()
		}
		if result.BaseResp.StatusCode == 0 && len(result.Choices) == 0 {
			return nil, ErrEmptyResponse
		}
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.TotalTokens
		msg := &schema.AIChatMessage{
			Content: result.Choices[0].Messages[0].Text,
		}
		if result.Choices[0].Messages[0].FunctionCall != nil {
			msg.FunctionCall = &schema.FunctionCall{
				Name:      result.Choices[0].Messages[0].FunctionCall.Name,
				Arguments: result.Choices[0].Messages[0].FunctionCall.Arguments,
			}
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
		o.usage = append(o.usage, result.Usage)
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

// CreateChat 支持全部参数的实现
func (o *Chat) CreateRawChat(ctx context.Context, r *minimaxclient2.CompletionRequest) (*minimaxclient2.Completion, error) {
	o.ResetUsage()
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, getPromptsFromMiniMaxMessageSets(r.Messages))
	}
	result, err := o.client.CreateCompletion(ctx, r)
	if err != nil {
		return nil, err
	}
	if result.InputSensitive {
		o.SetError(fmt.Sprintf("输入命中敏感词：%s", SensitiveTypeToValue(result.InputSensitiveType)))
		return nil, o.GetError()
	}
	if result.OutputSensitive {
		o.SetError(fmt.Sprintf("输出命中敏感词：%s", SensitiveTypeToValue(result.OutputSensitiveType)))
	}
	if result.BaseResp.StatusCode == 0 && len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	o.usage = append(o.usage, result.Usage)
	return result, err
}

func (o *Chat) GetNumTokens(text string) int {
	return 0
}

func (o *Chat) ResetUsage() {
	o.usage = make([]minimaxclient2.Usage, 0, 1)
}

func (o *Chat) SetError(text string) {
	o.chatError = errors.New(text)
}

func (o *Chat) GetError() error {
	return o.chatError
}

func (o *Chat) GetUsage() []minimaxclient2.Usage {
	return o.usage
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	return nil, errors.New("select embedding type")
}

// CreateEmbedding 存储文档向量化
func (o *Chat) CreateDbEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()

	result, err := o.client.CreateEmbedding(ctx, &minimaxclient2.EmbeddingPayload{
		Texts: inputTexts,
		Type:  "db", //db query
	})
	if err != nil {
		return nil, err
	}
	o.usage = append(o.usage, minimaxclient2.Usage{TotalTokens: result.TotalTokens})
	return result.Vectors, nil
}

// CreateQueryEmbedding 查询语句向量化
func (o *Chat) CreateQueryEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	o.ResetUsage()

	result, err := o.client.CreateEmbedding(ctx, &minimaxclient2.EmbeddingPayload{
		Texts: inputTexts,
		Type:  "query", //db query
	})
	if err != nil {
		return nil, err
	}
	o.usage = append(o.usage, minimaxclient2.Usage{TotalTokens: result.TotalTokens})
	return result.Vectors, nil
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

func getPromptsFromMiniMaxMessageSets(messageSets []*minimaxclient2.Message) []string {
	prompts := make([]string, 0, len(messageSets))
	for i := 0; i < len(messageSets); i++ {
		prompts = append(prompts, messageSets[i].Text)
	}

	return prompts
}

func messagesToClientMessages(messages []schema.ChatMessage) ([]*minimaxclient2.Message, minimaxclient2.BotSetting, minimaxclient2.ReplyConstraints) {

	setting := minimaxclient2.BotSetting{
		BotName: defaultBotName,
		Content: defaultBotDescription,
	}
	replyConstraints := minimaxclient2.ReplyConstraints{
		SenderType: defaultSendType,
		SenderName: defaultBotName,
	}
	msglen := len(messages)
	// 第一个system信息放入bot_setting
	if len(messages) > 0 {
		if messages[0].GetType() == schema.ChatMessageTypeSystem {
			setting.Content = messages[0].GetContent()
			messages = messages[1:]
			msglen -= 1
		}
	}

	msgs := make([]*minimaxclient2.Message, msglen)
	for i, m := range messages {
		typ := m.GetType()
		msg := &minimaxclient2.Message{
			SenderType: string(m.GetType()),
			SenderName: "",
			Text:       m.GetContent(),
		}

		switch typ {
		//case schema.ChatMessageTypeSystem:
		//	continue
		case schema.ChatMessageTypeAI:
			msg.SenderType = defaultSendType
			msg.SenderName = defaultBotName
		case schema.ChatMessageTypeHuman:
			msg.SenderType = "USER"
			msg.SenderName = defaultSendName
		case schema.ChatMessageTypeGeneric:
			msg.SenderType = "USER"
			msg.SenderName = defaultSendName
			//case schema.ChatMessageTypeFunction:
			//	msg.Role = "function"
		}
		msgs[i] = msg
	}

	return msgs, setting, replyConstraints
}
