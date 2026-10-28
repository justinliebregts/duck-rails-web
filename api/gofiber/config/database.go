package config

import (
	cfgMDL "gofiber/model/config"
	"os"
)

func LoadDBConfig() cfgMDL.DBConfig {
	return cfgMDL.DBConfig{
		User:     os.Getenv("DB_ACCOUNT"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Protocol: os.Getenv("DB_PROTOCOL"),
		DBName:   os.Getenv("DB_NAME"),
	}
}
