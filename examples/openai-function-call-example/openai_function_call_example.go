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
	baseUrl := "https://api.openai.com/v1"
	llm, err := openai.NewChat(openai.WithModel("gpt-4"), openai.WithBaseURL(baseUrl))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{Content: "查询机器人列表第二页的4条数据"},
	}, llms.WithFunctions(functions))
	if err != nil {
		log.Fatal(err)
	}

	if completion.FunctionCall != nil {
		fmt.Printf("Function call: %v\n", completion.FunctionCall)
	}
}

func getCurrentWeather(location string, unit string) (string, error) {
	weatherInfo := map[string]interface{}{
		"location":    location,
		"temperature": "72",
		"unit":        unit,
		"forecast":    []string{"sunny", "windy"},
	}
	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var functions = []llms.FunctionDefinition{
	{
		Name:        "get_shuiju_status", // name 必须为英文 ^[a-zA-Z0-9_-]{1,64}$
		Description: "根据用户传入的数量和页数，获取机器人详细信息的列表",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {"page": {"type": "string", "description": "页数，比如：1 就是指第1页"}, "size": {"type": "string","description":"数量，比如：10 就是看10条数据"}}, "required": ["page","size"]}`),
	},
}
