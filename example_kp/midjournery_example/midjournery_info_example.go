package main

import (
	"context"
	"example_kp/midjournery_example/internal/helpers"
	"fmt"
	"github.com/tmc/langchaingo/tgis/thenextleg"
	"os"
)

func main() {
	token := os.Getenv("MID_TOKEN")
	// 生成图片任务和查询进度  ----start
	c, err := thenextleg.New(thenextleg.WithAuthToken(token))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	resp, err := c.Info(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", resp)

	// get message

	msgId := resp.MessageId

	if msgId == "" {
		fmt.Println("message id is empty")
		return
	}
	respp, err := helpers.GetMessage(c, msgId)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", respp)
}
