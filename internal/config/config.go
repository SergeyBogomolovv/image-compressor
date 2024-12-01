package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr       string
	OutputPath string
}

func New() *Config {
	godotenv.Load()
	return &Config{
		Addr:       os.Getenv("ADDR"),
		OutputPath: os.Getenv("OUTPUT_PATH"),
	}
}
