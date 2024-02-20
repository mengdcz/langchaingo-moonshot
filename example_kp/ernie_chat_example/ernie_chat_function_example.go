package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/schema"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie"
)

func main() {

	//rdb := redis.NewClient(&redis.Options{
	//	Addr:     "192.168.5.89:6379",
	//	Password: "",
	//	DB:       0,
	//})
	k := os.Getenv("ERNIE_API_KEY")
	v := os.Getenv("ERNIE_SECRET_KEY")
	llm, err := ernie.NewChatWithCallback(callbacks.LogHandler{}, ernie.WithAKSK(k, v), ernie.WithModelName(ernie.ModelNameERNIEBot4))
	// note:
	// You would include ernie.WithAKSK(apiKey,secretKey) to use specific auth info.
	// You would include ernie.WithModelName(ernie.ModelNameERNIEBot) to use the ERNIE-Bot model.
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	/*	messages := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "你是靠谱Ai助手"},
		schema.HumanChatMessage{Content: "北京今天适合出游吗"},
	}*/
	fu := []llms.FunctionDefinition{
		{
			Name:        "get_current_weather",
			Description: "获取指定地点今天的天气",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]string{
						"type":        "string",
						"description": "省，市名，例如：河北省，石家庄",
					},
					"unit": map[string]any{
						"type": "string",
						"enum": []string{"摄氏度", "华氏度"},
					},
				},
				"required": []string{"location"},
			},
		},
	}
	/*	completion, err := llm.Call(ctx, messages,
			llms.WithTemperature(0.8),
			llms.WithFunctions(fu),
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				log.Println(string(chunk))
				return nil
			}),
		)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("=================")
		fmt.Printf("%#v\n", llm.GetUsage())
		fmt.Printf("%#v\n", completion.FunctionCall.Name)*/

	// 第二轮

	//{"name":"get_current_weather","thoughts":"用户询问的是关于天气的情况，需要调用get_current_weather工具来获取今天的天气信息。","arguments":"{\"location\":\"今天\"}"}

	messages2 := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "你是靠谱Ai助手"},
		schema.HumanChatMessage{Content: "今天适合出游吗"},
		schema.AIChatMessage{Content: "", FunctionCall: &schema.FunctionCall{
			Name:      "get_current_weather",
			Arguments: "{\"location\":\"北京\"}",
		}},
		schema.FunctionChatMessage{
			Name:    "get_current_weather",
			Content: "{\"temperature\": \"25\", \"unit\": \"摄氏度\", \"description\": \"晴朗\"}",
		},
	}
	completion2, err := llm.Call(ctx, messages2,
		llms.WithTemperature(0.8),
		llms.WithFunctions(fu),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			log.Println(string(chunk))
			return nil
		}),
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=================")
	fmt.Printf("%#v\n", llm.GetUsage())
	fmt.Printf("%#v\n", completion2)
}
