package adapter

import (
	"fmt"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Network  string
	Host     string
	Port     int
	Password string
	DB       int
}

type Redis struct {
	client *redis.Client
}

func NewRedisAdapter(cfg Config) *Redis {
	return &Redis{client: redis.NewClient(&redis.Options{
		Network:  cfg.Network,
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})}
}

func (a *Redis) GetClient() *redis.Client {
	return a.client
}
