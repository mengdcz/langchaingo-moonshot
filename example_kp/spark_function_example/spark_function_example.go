package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/spark"
	"github.com/tmc/langchaingo/schema"
	"log"
)

func main() {
	ctx := context.Background()

	// ===================== 单prompt ========================
	functions := []llms.FunctionDefinition{
		{
			Name:        "订单查询",
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
			Name:        "手机号归属地查询",
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
			Name:        "获取ai号列表",
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
			Name:        "税局状态查询",
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

	llm, err := spark.NewChat(spark.WithModel("spark3.0"))
	if err != nil {
		log.Fatal(err)
	}

	result, err := llm.Call(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{
			//Content: "我的手机号是18510775634，订单号是202310234567333，这个手机号是哪个区域的，并且订单内容是什么？",
			Content: "北京市税局状态",
		},
	},
		llms.WithFunctions(functions),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Println("stream .....")
			fmt.Println(string(chunk))
			fmt.Println(string("stream stop"))
			return nil
		}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	//// 函数名称
	//fmt.Println(result.FunctionCall.Name)
	//// 函数参数
	//fmt.Println(result.FunctionCall.Arguments)
	b, _ := json.Marshal(result)
	fmt.Println(string(b))
	fmt.Printf("%#v\n", llm.GetUsage())

}
