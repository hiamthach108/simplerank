package config

import (
	"os"

	"github.com/golobby/dotenv"
)

type AppConfig struct {
	App struct {
		Name    string `env:"APP_NAME"`
		Version string `env:"APP_VERSION"`
	}
	Server struct {
		Host string `env:"HTTP_HOST"`
		Port string `env:"HTTP_PORT"`
	}

	Cache struct {
		DefaultExpireTimeSec int    `env:"CACHE_DEFAULT_EXPIRE_TIME_SEC"`
		CleanupIntervalHour  int    `env:"CACHE_CLEANUP_INTERVAL_HOUR"`
		RedisHost            string `env:"REDIS_HOST"`
		RedisPort            string `env:"REDIS_PORT"`
		RedisPassword        string `env:"REDIS_PASSWORD"`
		RedisDB              int    `env:"REDIS_DB"`
	}
}

func NewAppConfig(envDir string) (*AppConfig, error) {
	config := &AppConfig{}
	file, err := os.Open(envDir)
	if err != nil {
		return nil, err
	}

	err = dotenv.NewDecoder(file).Decode(config)
	if err != nil {
		return config, err
	}

	return config, err
}
