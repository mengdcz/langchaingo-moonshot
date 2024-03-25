package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/moonshot"
	"github.com/tmc/langchaingo/schema"
	"os"
)

func main() {
	fmt.Println("moonshot start")
	ctx := context.Background()

	key := os.Getenv("MOONSHOT_API_KEY")
	model := "moonshot-v1-8k" // moonshot-v1-8k moonshot-v1-32k moonshot-v1-128k
	llm, err := moonshot.NewChat(
		moonshot.WithToken(key),
		moonshot.WithModel(model),
	)
	if err != nil {
		fmt.Printf(" moonshot.NewChat err \n", err.Error())
		return
	}
	messages := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "你是 M助理，由 M AI 提供的人工智能助手，你更擅长中文和英文的对话。你会为用户提供安全，有帮助，准确的回答。同时，你会拒绝一切涉及恐怖主义，种族歧视，黄色暴力等问题的回答。Moonshot AI 为专有名词，不可翻译成其他语言。"},
		schema.HumanChatMessage{Content: "介绍一下你自己,提供的人工智能助手，你更擅长中文和英文的对话。你会为用户提供安全，有帮助，准确的回答。同时，你会拒绝一切涉及恐怖主义，种族歧视，黄色暴力等问题的回答。Moonshot AI 为专有名词，不可翻译成其他语言"},
		schema.AIChatMessage{Content: "OK"},
		schema.HumanChatMessage{Content: "介绍一下你自己,提供的人工智能助手，你更擅长中文和英文的对话。你会为用户提供安全，有帮助，准确的回答。同时，你会拒绝一切涉及恐怖主义，种族歧视，黄色暴力等问题的回答。Moonshot AI 为专有名词，不可翻译成其他语言"},
	}

	temperature := 0.8
	completion, err := llm.Call(ctx,
		messages,
		llms.WithTemperature(temperature),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Println(string(chunk))
			return nil
		}),
	)
	if err != nil {
		fmt.Printf(" moonshot.NewChat  llm.Call err \n", err.Error())
	}

	fmt.Println("=================")
	fmt.Printf("%#v", llm.GetUsage())
	fmt.Println("=================")
	fmt.Printf("%#v", completion)
	// moonshot-v1-8k {PromptTokens:70, CompletionTokens:91, TotalTokens:161}}
	// moonshot-v1-128k {PromptTokens:70, CompletionTokens:86, TotalTokens:156}}

}
