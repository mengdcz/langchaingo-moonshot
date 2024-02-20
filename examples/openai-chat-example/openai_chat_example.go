package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	baseUrl := "https://apiagent.kaopuai.com/v1"
	ctx := context.Background()

	//llmCom, err := openai.New(openai.WithBaseURL(baseUrl), openai.WithModel("gpt-4"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//completion1, err1 := llmCom.Call(ctx, "今天天气怎么样")
	//if err1 != nil {
	//	fmt.Println(err1.Error())
	//	return
	//}
	//fmt.Println(completion1)
	//fmt.Println(llmCom.GetUsage())

	llm, err := openai.NewChat(openai.WithBaseURL(baseUrl), openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatal(err)
	}

	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.SystemChatMessage{Content: "你是一个AI助手"},
		schema.HumanChatMessage{Content: "今天天气怎么样"},
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	b, _ := json.Marshal(completion)
	fmt.Println(string(b))
	fmt.Println(llm.GetUsage())

}
