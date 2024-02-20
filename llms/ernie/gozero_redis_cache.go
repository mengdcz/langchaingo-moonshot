package ernie

import "github.com/zeromicro/go-zero/core/stores/redis"

type GozeroRedisCache struct {
	Rdb *redis.Redis
}

func (m GozeroRedisCache) Set(key string, value string, second int) error {
	err := m.Rdb.Set(key, value)
	if err != nil {
		return err
	}
	return m.Rdb.Expire(key, second)

}

func (m GozeroRedisCache) Get(key string) (string, error) {
	return m.Rdb.Get(key)
}
