package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/embeddings/emb_spark"
	"github.com/tmc/langchaingo/llms/spark"
	"log"
)

func main() {
	ctx := context.Background()

	//// ===================== 单prompt ========================
	//llm, err := spark.New()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//result, err := llm.Call(ctx, "介绍一下你自己")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println(result)
	//fmt.Printf("%#v\n", llm.GetUsage())
	//
	//// ===================== 多prompt ========================
	//res, err := llm.Generate(ctx, []string{"介绍一下你自己", "介绍一下刘德华"})
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println(res[0].Text)
	//fmt.Println(res[1].Text)
	//fmt.Printf("%#v\n", llm.GetUsage())

	//// ===================== chat ========================
	//llmChat, err := spark.NewChat()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//messages := []schema.ChatMessage{
	//	schema.HumanChatMessage{Content: "介绍一下你自己"},
	//}
	//completion, err := llmChat.Call(ctx, messages,
	//	llms.WithTemperature(0.8),
	//	llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
	//		log.Println(string(chunk))
	//		return nil
	//	}),
	//)
	//
	//if err != nil {
	//	log.Println("error")
	//	log.Fatal(err)
	//}
	//
	//// 同一个llmChat对象 并发处理回复时，可能会导致GetUsage方法不准确
	//log.Printf("%#v\n", llmChat.GetUsage())
	//
	//fmt.Printf("%v\n", completion)

	// ===================== 向量 ========================
	llmEmb, err := spark.New()
	if err != nil {
		log.Fatal(err)
	}
	emb, err := emb_spark.NewSpark(emb_spark.WithClient(*llmEmb))
	if err != nil {
		log.Fatal(err)
	}
	resemb, err := emb.EmbedQuery(ctx, "高考成绩")
	if err != nil {
		log.Println(err.Error())
		log.Fatal("emb.EmbedDocuments(ctx, []string{\"靠谱前程\", \"高考成绩\"}) error")
	}
	fmt.Println(resemb)
}
