package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/minimax"
	"github.com/tmc/langchaingo/llms/minimax/minimaxclient"
)

func main() {

	ctx := context.Background()
	llmmini, err := minimax.NewChat()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	r := &minimaxclient.CompletionRequest{
		Messages: []*minimaxclient.Message{{
			SenderType: "USER",
			SenderName: "用户",
			Text:       "使用英语翻译这句话：我是谁",
		}},
		BotSetting: []minimaxclient.BotSetting{
			{
				BotName: "靠谱AI智能助手",
				Content: "你是靠谱AI智能助手",
			},
		},
		ReplyConstraints: minimaxclient.ReplyConstraints{
			SenderType: "BOT",
			SenderName: "靠谱AI智能助手",
			Glyph: &minimaxclient.Glyph{
				Type:     "raw",
				RawGlyph: "{{gen 'content'}}",
			},
		},
		//StreamingFunc:       nil,
		//SampleMessages:      nil,
		//Functions:           nil,
		//FunctionCallSetting: nil,
		//Plugins:             nil,
	}
	result, err := llmmini.CreateRawChat(ctx, r)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("completion:%s\n", result.Reply)
	fmt.Println(llmmini.GetUsage())
}
