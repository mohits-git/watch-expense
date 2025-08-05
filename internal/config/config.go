package config

type Config struct {
	Port string
	DSN  string
	// and more fields...
}

func Load() *Config {
	// TODO: Load configuration from environment variables, files, etc.
	// For simplicity, returning a hardcoded config.
	return &Config{
		Port: "8080",
		DSN:  "user:password@tcp(localhost:3306)/dbname",
	}
}
