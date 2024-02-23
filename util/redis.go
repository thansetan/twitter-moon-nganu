package util

import "github.com/redis/go-redis/v9"

func NewRedisClient(redisURL string) (*redis.Client, error) {
	url, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	redis := redis.NewClient(url)
	return redis, nil
}
