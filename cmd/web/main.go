package main

import (
	"context"
	"log"
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/bcrypt"
	"github.com/mohits-git/watch-expense/internal/adapters/http/handlers"
	"github.com/mohits-git/watch-expense/internal/adapters/http/router"
	"github.com/mohits-git/watch-expense/internal/adapters/imageupload"
	"github.com/mohits-git/watch-expense/internal/adapters/jwttoken"
	"github.com/mohits-git/watch-expense/internal/adapters/mysql"
	"github.com/mohits-git/watch-expense/internal/services"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg := LoadConfig()

	// Setup database
	db := SetupDB(ctx, cfg.MYSQL_DSN)
	defer db.Close()

	// utils
	bcryptProvidder := bcrypt.NewBcryptPasswordHasher(12)
	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)

	// repositories
	userRepo := mysql.NewUserRepository(db)
	departmentRepo := mysql.NewDepartmentRepository(db)
	projectRepo := mysql.NewProjectRepository(db)
	expenseRepo := mysql.NewExpenseRepository(db)
	advanceRepo := mysql.NewAdvanceRepository(db)

	// services
	authService := services.NewAuthenticationService(userRepo, tokenProvider, bcryptProvidder)
	userService := services.NewUserService(userRepo)
	departmentService := services.NewDepartmentService(departmentRepo)
	projectService := services.NewProjectService(projectRepo)
	expenseService := services.NewExpenseService(expenseRepo)
	advanceService := services.NewAdvanceService(advanceRepo)
	imageService := imageupload.NewFSImageUpload("http://localhost:8080", cfg.UPLOAD_DIRECTORY)

	// handlers
	commonHandler := handlers.NewCommonHandler()
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	projectHandler := handlers.NewProjectHandler(projectService)
	departmentHandler := handlers.NewDepartmentHandler(departmentService)
	expenseHandler := handlers.NewExpenseHandler(expenseService)
	advanceHandler := handlers.NewAdvanceHandler(advanceService)
	imageHandler := handlers.NewImageUploadHandler(imageService)

	// middleware
	authMiddleware := handlers.NewAuthMiddleware(tokenProvider)

	// router
	httpRouter := router.NewHTTPRouter(
		authMiddleware,
		commonHandler,
		authHandler,
		userHandler,
		projectHandler,
		departmentHandler,
		expenseHandler,
		advanceHandler,
		imageHandler,
	)

	// Start the HTTP server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", enableCors(httpRouter)))
}

// http.Handler wrapper that adds CORS headers to responses.
func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins, or specify a list
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // If you need to send cookies/auth headers

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
