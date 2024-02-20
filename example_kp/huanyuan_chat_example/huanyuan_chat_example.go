package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/hunyuan"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := hunyuan.New()
	if err != nil {
		fmt.Println(err.Error())
	}
	promot := "介绍一下你自己"
	resp, err := llm.Call(context.Background(), promot)
	if err != nil {
		fmt.Println(err.Error())
	}
	b, _ := json.Marshal(resp)
	fmt.Println("completion resp: ", string(b))
	fmt.Println("completion usage: ", llm.GetUsage())

	llmChat, err := hunyuan.NewChat()
	if err != nil {
		fmt.Println(err.Error())
	}
	messages := []schema.ChatMessage{
		schema.HumanChatMessage{Content: "介绍一下你自己"},
	}
	resp1, err := llmChat.Call(context.Background(), messages,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Println(string(chunk))
			return nil
		}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("chat stream resp1: ", resp1)
	fmt.Println("chat stream uage:", llmChat.GetUsage())
}
