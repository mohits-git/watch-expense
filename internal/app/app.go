package app

import "github.com/mohits-git/watch-expense/internal/config"

type App struct {
	config *config.Config
	// and more fields... db, logger, etc.
}

func NewApp(config *config.Config) *App {
	return &App{
		config: config,
		// Initialize other fields like db, logger, etc.
	}
}
