package main

import (
	"log"
	"net/http"

	"github.com/mohits-git/watch-expense/internal/app"
	"github.com/mohits-git/watch-expense/internal/config"
	"github.com/mohits-git/watch-expense/internal/router"
)

func main() {
	// Load configuration
	cfg := config.Load()
	// Initialize the application
	app := app.NewApp(cfg)
	// Initialize the router with the app instance
	route := router.NewRouter(app).Route()
	// Start the HTTP server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", route))
}
