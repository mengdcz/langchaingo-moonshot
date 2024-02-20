package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/minimax"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	ctx := context.Background()

	llmChat, err := minimax.NewChat()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	messages := []schema.ChatMessage{
		schema.SystemChatMessage{
			Content: "我是AI智能助理",
		},
		schema.HumanChatMessage{Content: "介绍一下你自己"},
	}
	result, err := llmChat.Call(ctx, messages, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Println("=======", string(chunk))
		return nil
	}))
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("result", result.Content)
	fmt.Println("usage", llmChat.GetUsage())

}
