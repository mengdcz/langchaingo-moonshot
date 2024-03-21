package main

import (
	"context"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"os"

	// "github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie"
	"github.com/tmc/langchaingo/schema"
	"log"
)

func main() {

	goredisDb := goredis.NewClient(&goredis.Options{
		Addr:     "192.168.5.89:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cache := ernie.GoRedisCache{
		Rdb: goredisDb,
	}

	//conf := redis.RedisConf{
	//	Host:        "192.168.5.89:6379",
	//	Type:        "node",
	//	Pass:        "",
	//	Tls:         false,
	//	NonBlock:    false,
	//	PingTimeout: time.Second,
	//}
	//rdb := redis.MustNewRedis(conf)
	//
	//cache := ernie.GozeroRedisCache{
	//	Rdb: rdb,
	//}
	k := os.Getenv("ERNIE_API_KEY")
	v := os.Getenv("ERNIE_SECRET_KEY")

	llm, err := ernie.NewChatWithCallback(callbacks.LogHandler{}, ernie.WithAKSK(k, v), ernie.WithModelName(ernie.ModelNameERNIEBot4), ernie.WithCache(cache))
	// note:
	// You would include ernie.WithAKSK(apiKey,secretKey) to use specific auth info.
	// You would include ernie.WithModelName(ernie.ModelNameERNIEBot) to use the ERNIE-Bot model.
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	messages := []schema.ChatMessage{
		schema.HumanChatMessage{Content: "介绍一下你自己"},
	}
	completion, err := llm.Call(ctx, messages,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			log.Println(string(chunk))
			return nil
		}),
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=================")
	fmt.Printf("%#v", llm.GetUsage())
	fmt.Printf("%#v", completion)

}
