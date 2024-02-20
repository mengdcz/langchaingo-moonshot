package main

import (
	"context"
	"encoding/json"
	"fmt"
	ernieembedding "github.com/tmc/langchaingo/embeddings/ernie"
	"github.com/tmc/langchaingo/llms/ernie"
	"golang.org/x/sys/windows"
	"os"
)

func main() {
	k := os.Getenv("ERNIE_API_KEY")
	v := os.Getenv("ERNIE_SECRET_KEY")
	c, err := ernie.New(ernie.WithAKSK(k, v))
	if err != nil {
		fmt.Println(err)
		return
	}
	emb, err := ernieembedding.NewErnie(ernieembedding.WithClient(*c))
	if err != nil {
		fmt.Println(err)
		return
	}
	content := "曹云金口碑为什么会发生反转"
	res, err := emb.EmbedQuery(context.Background(), content)
	if err != nil {
		fmt.Println(err)
		return
	}
	a := map[string]interface{}{
		"vector": res,
	}
	aa, _ := json.Marshal(a)
	os.WriteFile("./emb_query.txt", aa, windows.FILE_ACTION_ADDED)
}
