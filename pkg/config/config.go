package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port     int    `env:"SERVER_PORT" envDefault:"10010"`
	Env      string `env:"SERVER_ENV" envDefault:"dev"`
	DBHost   string `env:"DB_HOST" envDefault:"localhost"`
	DBPort   int    `env:"DB_PORT" envDefault:"10001"`
	DBUser   string `env:"DB_USER"`
	DBPass   string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	DBURL    string
	RedisAddr string `env:"REDIS_ADDR"`
	RedisPass string `env:"REDIS_PASS"`
	S3Bucket string `env:"S3_BUCKET"`
	S3Region string `env:"S3_REGION"`
	S3Access string `env:"S3_ACCESS_KEY"`
	S3Secret string `env:"S3_SECRET_KEY"`
	S3End    string `env:"S3_ENDPOINT" envDefault:"is3.cloudhost.id"`
	JwtSecret string `env:"JWT_SECRET" envDefault:"utschool"`
}

var cfg *Config

func LoadConfig() *Config {
	var c Config
	_ = godotenv.Load() // ini akan load file .env ke environment
	if err := env.Parse(&c); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	// If DBURL is not explicitly set, construct it from individual DB components
	// This ensures that DB_HOST and DB_PORT from .env are respected
	if c.DBURL == "" {
		c.DBURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
	}

	cfg = &c
	return cfg
}

func GetConfig() *Config {
	if cfg == nil {
		log.Fatal("Config is not loaded yet. Call LoadConfig first.")
	}
	return cfg
}
