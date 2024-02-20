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

	// =============function call
	messages1 := []schema.ChatMessage{
		schema.SystemChatMessage{
			Content: "我是AI智能助理",
		},
		schema.HumanChatMessage{Content: "北京市税局状态"},
	}
	functions := []llms.FunctionDefinition{
		{
			Name:        "order_query",
			Description: "根据订单号查询订单详情",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"order_no": map[string]string{
						"type":        "string",
						"description": "订单号，15位的数字，比如：201110112222333",
					},
				},
				"required": []string{"order_no"},
			},
		},
		{
			Name:        "mobile_address",
			Description: "根据手机号查询手机号所属地区",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mobile_no": map[string]string{
						"type":        "string",
						"description": "手机号，11位的数字，比如：18510771234",
					},
				},
				"required": []string{"mobile_no"},
			},
		},
		{
			Name:        "get_ai_list",
			Description: "获取ai号列表可以提供ai号相关信息，你可以提供指定的页数，指定的数量来检索ai号信息",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{
					//"page_no": map[string]string{
					//	"type":        "string",
					//	"description": "手机号，11位的数字，比如：18510771234",
					//},
				},
				//"required": []string{"mobile_no"},
			},
		},
		{
			Name:        "shuiju_status",
			Description: "根据输入的地区获取当地的税务局状态信息",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"province": map[string]string{
						"type":        "string",
						"description": "地区，省份，例如：湖北省",
					},
				},
				"required": []string{"province"},
			},
		},
	}

	result1, err := llmChat.Call(ctx, messages1, llms.WithFunctions(functions),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {

			fmt.Println("chunk", string(chunk))
			return nil
		}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("result", result1.FunctionCall)
	fmt.Println("usage", llmChat.GetUsage())
}
