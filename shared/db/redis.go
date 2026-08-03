package db

import "github.com/redis/go-redis/v9"

func NewRedis(uri string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     uri,
		DB:       0,
		Password: "",
	})
	return rdb
}
