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
		// - image upload routes
		"GET /public/images/": http.StripPrefix("/public/images/", http.FileServer(http.Dir("./public/images"))).ServeHTTP,
		"POST /api/images":    imageUploadHandler.HandleUploadImage,
		"DELETE /api/images":  imageUploadHandler.HandleDeleteImage,
	}

	authenticatedRoutes := map[string]http.HandlerFunc{
		// - auth routes
		"GET /api/auth/me": authHandler.HandleMe,
		// - user routes
		"GET /api/users":        userHandler.HandleGetAllUsers,
		"GET /api/users/{id}":   userHandler.HandleGetUserByID,
		"POST /api/users":       userHandler.HandleCreateUser,
		"PUT /api/users/{id}":   userHandler.HandleUpdateUser,
		"GET /api/users/budget": userHandler.HandleGetUserBudget,
		// - project routes
		"GET /api/admin/projects":      projectHandler.HandleGetAllProjects,
		"GET /api/admin/projects/{id}": projectHandler.HandleGetProjectByID,
		"POST /api/admin/projects":     projectHandler.HandleCreateProject,
		"PUT /api/admin/projects/{id}": projectHandler.HandleUpdateProject,
		// - department routes
		"GET /api/admin/departments":      departmentHandler.HandleGetAllDepartments,
		"GET /api/admin/departments/{id}": departmentHandler.HandleGetDepartmentByID,
		"POST /api/admin/departments":     departmentHandler.HandleCreateDepartment,
		"PUT /api/admin/departments/{id}": departmentHandler.HandleUpdateDepartment,
		// - expenses routes
		"POST /api/expenses":        expenseHandler.HandleCreateExpense,
		"GET /api/expenses":         expenseHandler.HandleGetExpenses,
		"GET /api/expenses/{id}":    expenseHandler.HandleGetExpenseByID,
		"PUT /api/expenses/{id}":    expenseHandler.HandleUpdateExpense,
		"PATCH /api/expenses/{id}":  expenseHandler.HandleUpdateExpenseStatus,
		"GET /api/expenses/summary": expenseHandler.HandleGetExpenseSummary,
		// - advance routes
		"POST /api/advance-request":        advanceHandler.HandleCreateAdvance,
		"GET /api/advance-request":         advanceHandler.HandleGetAdvances,
		"GET /api/advance-request/{id}":    advanceHandler.HandleGetAdvanceByID,
		"PUT /api/advance-request/{id}":    advanceHandler.HandleUpdateAdvance,
		"PATCH /api/advance-request/{id}":  advanceHandler.HandleUpdateAdvanceStatus,
		"GET /api/advance-request/summary": advanceHandler.HandleGetAdvanceSummary,
	}

	for route, handler := range publicRoutes {
		mux.HandleFunc(route, handler)
	}

	for route, handler := range authenticatedRoutes {
		mux.HandleFunc(route, authMiddleware.Authenticated(handler))
	}

	return mux
}
