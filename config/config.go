package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBURL      string

	// Server
	SERVER_ADDRESS      string
	TEST_SERVER_ADDRESS string
	EXTERNAL_API_URL    string
	LOG_LEVEL           string
}

func LoadEnv() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed loading .env file: %v", err)
	}

	return &Config{
		DBHost:              os.Getenv("DB_HOST"),
		DBPort:              os.Getenv("DB_PORT"),
		DBUser:              os.Getenv("DB_USER"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		DBURL:               os.Getenv("DATABASE_URL"),
		SERVER_ADDRESS:      os.Getenv("SERVER_ADDRESS"),
		TEST_SERVER_ADDRESS: os.Getenv("TEST_SERVER_ADDRESS"),
		EXTERNAL_API_URL:    os.Getenv("EXTERNAL_API_URL"),
		LOG_LEVEL:           os.Getenv("LOG_LEVEL"),
	}, nil
}
