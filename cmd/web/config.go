package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ENVIRONMENT      string
	MYSQL_DSN        string
	JWT_SECRET       string
	JWT_ISSUER       string
	JWT_AUDIENCE     string
	UPLOAD_DIRECTORY string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var config Config

	config.ENVIRONMENT = os.Getenv("ENVIRONMENT")
	if config.ENVIRONMENT == "" {
		config.ENVIRONMENT = "development"
	}

	config.MYSQL_DSN = os.Getenv("MYSQL_DSN")
	if config.MYSQL_DSN == "" {
		config.MYSQL_DSN = "root:mysecretpassword@tcp(localhost:3306)/watchexpense"
	}

	config.JWT_SECRET = os.Getenv("JWT_SECRET")
	if config.JWT_SECRET == "" {
		config.JWT_SECRET = "jwt_secret_pass"
	}

	config.JWT_ISSUER = os.Getenv("JWT_ISSUER")
	if config.JWT_ISSUER == "" {
		config.JWT_ISSUER = "jwt_issuer_name"
	}

	config.JWT_AUDIENCE = os.Getenv("JWT_AUDIENCE")
	if config.JWT_AUDIENCE == "" {
		config.JWT_AUDIENCE = "jwt_audience"
	}

	config.UPLOAD_DIRECTORY = os.Getenv("UPLOAD_DIRECTORY")
	if config.UPLOAD_DIRECTORY == "" {
		config.UPLOAD_DIRECTORY = "./public/images"
	}

	return config
}
