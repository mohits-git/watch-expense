package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ENVIRONMENT      string
	DYNAMODB_TABLE   string
	JWT_SECRET       string
	JWT_ISSUER       string
	JWT_AUDIENCE     string
	UPLOAD_DIRECTORY string
	S3_BUCKET_NAME   string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	var config Config

	config.ENVIRONMENT = os.Getenv("ENVIRONMENT")
	if config.ENVIRONMENT == "" {
		config.ENVIRONMENT = "production"
	}

	config.DYNAMODB_TABLE = os.Getenv("DYNAMODB_TABLE")
	if config.DYNAMODB_TABLE == "" {
		config.DYNAMODB_TABLE = "watch-expense-table"
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

	config.S3_BUCKET_NAME = os.Getenv("S3_BUCKET_NAME")
	if config.S3_BUCKET_NAME == "" {
		config.S3_BUCKET_NAME = "watch-expense-bucket"
	}

	return config
}
