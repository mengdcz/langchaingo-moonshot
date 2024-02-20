package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/chatglm"
	"github.com/tmc/langchaingo/schema"
	"log"
)

func main() {
	ctx := context.Background()

	//llm, err := chatglm.New(chatglm.WithEnableSearch(true), chatglm.WithSearchQuery(""))
	//if err != nil {
	//	log.Fatal(err)
	//}

	//result, err := llm.Call(ctx, "go语言写一个冒泡")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println(result)
	//fmt.Printf("%#v\n", llm.GetUsage())
	//path := "./aaa.txt"
	//err1 := os.WriteFile(path, []byte(result), 0666)
	//if err1 != nil {
	//	fmt.Println(err1.Error())
	//}
	//fmt.Println("write success")

	//model := "chatglm_turbo"
	model := "characterglm"
	llmChat, err := chatglm.NewChat(chatglm.WithModel(model))
	if err != nil {
		log.Fatal(err)
	}
	messages := []schema.ChatMessage{
		schema.HumanChatMessage{Content: "介绍一下你自己"},
	}

	completion, err := llmChat.Call(ctx, messages,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			log.Println(string(chunk))
			return nil
		}),
	)

	if err != nil {
		log.Println("error")
		log.Fatal(err)
	}

	// 同一个llmChat对象 并发处理回复时，可能会导致GetUsage方法不准确
	log.Printf("%#v\n", llmChat.GetUsage())

	fmt.Printf("%v\n", completion)

	// 向量
	//emb, err := emb_chatglm.NewChatglm(emb_chatglm.WithClient(*llm))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//res, err := emb.EmbedDocuments(ctx, []string{"靠谱前程", "高考成绩"})
	//if err != nil {
	//	log.Println(err.Error())
	//	log.Fatal("emb.EmbedDocuments(ctx, []string{\"靠谱前程\", \"高考成绩\"}) error")
	//}
	//fmt.Println(res)
}
