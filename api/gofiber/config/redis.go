package config

import (
	cfgMDL "gofiber/model/config"
	"os"
	"strconv"
)

func LoadCacheConfig() cfgMDL.CacheConfig {

	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))

	return cfgMDL.CacheConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     redisPort,
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

}
