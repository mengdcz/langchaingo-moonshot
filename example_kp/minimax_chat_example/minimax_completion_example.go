package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/minimax"
)

func main() {

	llmmini, err := minimax.New()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	result, err := llmmini.Call(context.Background(), "介绍一下你自己")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("completion:%s\n", result)
	fmt.Println(llmmini.GetUsage())

}
