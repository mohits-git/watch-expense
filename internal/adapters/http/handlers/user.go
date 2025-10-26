package handlers

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService}
}

func (h *UserHandler) HandleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
	}
	responseUsers := []dtos.User{}
	for _, user := range users {
		responseUsers = append(responseUsers, dtos.ToUserDTO(user))
	}
	writeResponse(w, http.StatusOK, "users fetched successfully", responseUsers)
}

func (h *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "user not found")
		} else if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid user ID")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	responseUser := dtos.ToUserDTO(user)
	writeResponse(w, http.StatusOK, "user fetched successfully", responseUser)
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	createUserRequest, err := decodeRequest[dtos.CreateUserRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user := dtos.ToUserDomain(dtos.User{
		EmployeeId:   createUserRequest.EmployeeId,
		Name:         createUserRequest.Name,
		Password:     createUserRequest.Password,
		Email:        createUserRequest.Email,
		Role:         createUserRequest.Role,
		ProjectID:    createUserRequest.ProjectID,
		DepartmentID: createUserRequest.DepartmentID,
	})
	userID, err := h.userService.CreateUser(r.Context(), user)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "only admin can create users")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid user data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusCreated, "user created successfully", dtos.CreateUserResponse{ID: userID})
}

func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	updateUserRequest, err := decodeRequest[dtos.UpdateUserRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user := dtos.ToUserDomain(dtos.User{
		EmployeeId:   updateUserRequest.EmployeeId,
		Name:         updateUserRequest.Name,
		Password:     updateUserRequest.Password,
		Email:        updateUserRequest.Email,
		Role:         updateUserRequest.Role,
		ProjectID:    updateUserRequest.ProjectID,
		DepartmentID: updateUserRequest.DepartmentID,
	})
	user.ID = userID

	err = h.userService.UpdateUser(r.Context(), user)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "only admin can update users")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid user data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusOK, "user updated successfully", struct{}{})
}

func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "only admin can delete users")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid user ID")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusOK, "user deleted successfully", struct{}{})
}

func (h *UserHandler) HandleGetUserBudget(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, "success", dtos.GetUserBudgetResponse{
		Budget: 289,
	})
	// TODO:
	// userID := r.URL.Query().Get("userId")
	// budget, err := h.userService.GetUserBudget(r.Context(), userID)
	// if err != nil {
	//   if apperr.IsUnauthorizedError(err) {
	//     writeError(w, http.StatusUnauthorized, "unauthorized")
	//   } else if apperr.IsNotFoundError(err) {
	//     writeError(w, http.StatusNotFound, "user not found")
	//   } else if apperr.IsInvalidError(err) {
	//     writeError(w, http.StatusBadRequest, "invalid user ID")
	//   } else {
	//     writeError(w, http.StatusInternalServerError, "internal server error")
	//   }
	//   return
	// }
	// writeResponse(w, http.StatusOK, "user budget fetched successfully", dtos.GetUserBudgetResponse{Budget: budget})
}
