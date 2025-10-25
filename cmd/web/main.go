package main

import (
	"context"
	"log"
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/handlers"
	"github.com/mohits-git/watch-expense/internal/adapters/http/router"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg := LoadConfig()

	// Setup database
	db := SetupDB(ctx, cfg.MYSQL_DSN)
	defer db.Close()

	// handlers
	commonHandler := handlers.NewCommonHandler()

	// router
	httpRouter := router.NewHTTPRouter(commonHandler)

	// Start the HTTP server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", httpRouter))
}
