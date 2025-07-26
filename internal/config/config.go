package config

import (
	"house-of-neural-networks/internal/triton"
	"house-of-neural-networks/pkg/db/cache"
	"house-of-neural-networks/pkg/db/postgres"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	postgres.Config
	cache.RedisConfig
	triton.TritonConfig

	GRPCServerPort int    `env:"GRPC_SERVER_PORT" env-default:"50051"`
	JWTSecret      string `env:"JWT_SECRET" env-default:""`

	// For Gateway
	HTTPServerPort    int    `env:"HTTP_SERVER_PORT" env-default:"8080"`
	AuthServiceURL    string `env:"AUTH_SERVICE_URL" env-default:"localhost:50051"`
	ModelServiceURL   string `env:"MODEL_SERVICE_URL" env-default:"localhost:50052"`
	MessageServiceURL string `env:"MESSAGE_SERVICE_URL" env-default:"localhost:50053"`
}

func New() *Config {
	cfg := Config{}
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil
	}
	return &cfg
}
