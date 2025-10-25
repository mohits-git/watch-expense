package router

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/handlers"
)

func NewHTTPRouter(
	authMiddleware *handlers.AuthMiddleware,
	commonHandler *handlers.CommonHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	departmentHandler *handlers.DepartmentHandler,
	expenseHandler *handlers.ExpenseHandler,
	advanceHandler *handlers.AdvanceHandler,
	imageUploadHandler *handlers.ImageUploadHandler,
) http.Handler {
	mux := http.NewServeMux()

	publicRoutes := map[string]http.HandlerFunc{
		// - docs routes
		"GET /docs/openapi.yaml": commonHandler.ServeOpenAPISpec,
		"GET /docs/":             commonHandler.SwaggerUIHandler().ServeHTTP,
		// - health check
		"GET /health": commonHandler.HealthCheck,
		// - auth routes
		"POST /api/auth/login":  authHandler.HandleLogin,
		"POST /api/auth/logout": authMiddleware.WithToken(authHandler.HandleLogout),
	}

	authenticatedRoutes := map[string]http.HandlerFunc{
		// - auth routes
		"GET /api/auth/me": authHandler.HandleMe,
		// - user routes
		"GET /api/users":      userHandler.HandleGetAllUsers,
		"GET /api/users/{id}": userHandler.HandleGetUserByID,
		"POST /api/users":     userHandler.HandleCreateUser,
		"PUT /api/users/{id}": userHandler.HandleUpdateUser,
		// - project routes
		"GET /api/projects":      projectHandler.HandleGetAllProjects,
		"GET /api/projects/{id}": projectHandler.HandleGetProjectByID,
		"POST /api/projects":     projectHandler.HandleCreateProject,
		"PUT /api/projects/{id}": projectHandler.HandleUpdateProject,
		// - department routes
		"GET /api/departments":      departmentHandler.HandleGetAllDepartments,
		"GET /api/departments/{id}": departmentHandler.HandleGetDepartmentByID,
		"POST /api/departments":     departmentHandler.HandleCreateDepartment,
		"PUT /api/departments/{id}": departmentHandler.HandleUpdateDepartment,
		// - expenses routes
		"POST /api/expenses":       expenseHandler.HandleCreateExpense,
		"GET /api/expenses":        expenseHandler.HandleGetExpenses,
		"GET /api/expenses/{id}":   expenseHandler.HandleGetExpenseByID,
		"PUT /api/expenses/{id}":   expenseHandler.HandleUpdateExpense,
		"PATCH /api/expenses/{id}": expenseHandler.HandleUpdateExpenseStatus,
		// - advance routes
		"POST /api/advances":       advanceHandler.HandleCreateAdvance,
		"GET /api/advances":        advanceHandler.HandleGetAdvances,
		"GET /api/advances/{id}":   advanceHandler.HandleGetAdvanceByID,
		"PUT /api/advances/{id}":   advanceHandler.HandleUpdateAdvance,
		"PATCH /api/advances/{id}": advanceHandler.HandleUpdateAdvanceStatus,
		// - image upload routes
		"POST /api/images":   imageUploadHandler.HandleUploadImage,
		"DELETE /api/images": imageUploadHandler.HandleDeleteImage,
	}

	for route, handler := range publicRoutes {
		mux.HandleFunc(route, handler)
	}

	for route, handler := range authenticatedRoutes {
		mux.HandleFunc(route, authMiddleware.Authenticated(handler))
	}

	return mux
}
