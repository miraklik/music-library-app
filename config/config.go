package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBURL               string
	SERVER_ADDRESS      string
	TEST_SERVER_ADDRESS string
}

func LoadEnv() *Config {
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
	}
}
