package ernie

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type GoRedisCache struct {
	Rdb *redis.Client
}

func (g GoRedisCache) Set(key string, value string, second int) error {
	err := g.Rdb.Set(context.Background(), key, value, time.Duration(second)*time.Second).Err()
	if err != nil {
		fmt.Println("goRdisCache error set:", err.Error())
	}
	return err
}

func (g GoRedisCache) Get(key string) (string, error) {
	val, err := g.Rdb.Get(context.Background(), key).Result()
	if err != nil {
		fmt.Println("goRedisCache error get ", err.Error())
	}
	return val, err
}
