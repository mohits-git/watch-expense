package handlers

import (
	"net/http"
	"strconv"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type ExpenseHandler struct {
	expenseService services.ExpenseService
}

func NewExpenseHandler(expenseService services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService}
}

func (h *ExpenseHandler) HandleCreateExpense(w http.ResponseWriter, r *http.Request) {
	createExpenseRequest, err := decodeRequest[dtos.CreateExpenseRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	expense := dtos.ToCreateExpenseDomain(createExpenseRequest)
	id, err := h.expenseService.CreateExpense(r.Context(), expense)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid expense data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	createExpenseResponse := dtos.CreateExpenseResponse{ID: id}
	writeResponse(w, http.StatusCreated, "expense created successfully", createExpenseResponse)
}

func (h *ExpenseHandler) HandleGetExpenses(w http.ResponseWriter, r *http.Request) {
	filterOptions, err := parseExpensesFilterOptions(r)
	if err != nil {
		if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid query parameters")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
	}
	expenses, total, err := h.expenseService.GetAllExpenses(r.Context(), filterOptions)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	expenseDtos := []dtos.Expense{}
	for _, expense := range expenses {
		expenseDtos = append(expenseDtos, dtos.ToExpenseDTO(expense))
	}
	getExpensesResponse := dtos.GetExpensesResponse{
		TotalExpenses: total,
		Expenses:      expenseDtos,
	}
	writeResponse(w, http.StatusOK, "expenses fetched successfully", getExpensesResponse)
}

func (h *ExpenseHandler) HandleUpdateExpense(w http.ResponseWriter, r *http.Request) {
  updateExpenseRequest, err := decodeRequest[dtos.UpdateExpenseRequest](r)
  if err != nil {
    writeError(w, http.StatusBadRequest, "invalid request")
    return
  }
  expenseID := r.PathValue("id")
  expense := domain.Expense{
    ID:           expenseID,
    Amount:       updateExpenseRequest.Amount,
    Description:  updateExpenseRequest.Description,
    Purpose:      updateExpenseRequest.Purpose,
    IsReconciled: updateExpenseRequest.IsReconciled,
  }
  err = h.expenseService.UpdateExpense(r.Context(), expense)
  if err != nil {
    if apperr.IsUnauthorizedError(err) {
      writeError(w, http.StatusUnauthorized, "unauthorized")
    } else if apperr.IsForbiddenError(err) {
      writeError(w, http.StatusForbidden, "forbidden")
    } else if apperr.IsNotFoundError(err) {
      writeError(w, http.StatusNotFound, "expense not found")
    } else if apperr.IsInvalidError(err) {
      writeError(w, http.StatusBadRequest, "invalid expense data")
    } else {
      writeError(w, http.StatusInternalServerError, "internal server error")
    }
    return
  }
  writeResponse(w, http.StatusOK, "expense updated successfully", struct{}{})
}

func (h *ExpenseHandler) HandleUpdateExpenseStatus(w http.ResponseWriter, r *http.Request) {
  updateExpenseStatusRequest, err := decodeRequest[dtos.UpdateExpenseStatusRequest](r)
  if err != nil {
    writeError(w, http.StatusBadRequest, "invalid request")
    return
  }
  expenseID := r.PathValue("id")
  err = h.expenseService.UpdateExpenseStatus(r.Context(), expenseID, updateExpenseStatusRequest.Status)
  if err != nil {
    if apperr.IsUnauthorizedError(err) {
      writeError(w, http.StatusUnauthorized, "unauthorized")
    } else if apperr.IsForbiddenError(err) {
      writeError(w, http.StatusForbidden, "forbidden")
    } else if apperr.IsNotFoundError(err) {
      writeError(w, http.StatusNotFound, "expense not found")
    } else if apperr.IsInvalidError(err) {
      writeError(w, http.StatusBadRequest, "invalid status")
    } else {
      writeError(w, http.StatusInternalServerError, "internal server error")
    }
    return
  }
  writeResponse(w, http.StatusOK, "expense status updated successfully", struct{}{})
}

func (h *ExpenseHandler) HandleGetExpenseByID(w http.ResponseWriter, r *http.Request) {
  expenseID := r.PathValue("id")
  expense, err := h.expenseService.GetExpenseByID(r.Context(), expenseID)
  if err != nil {
    if apperr.IsUnauthorizedError(err) {
      writeError(w, http.StatusUnauthorized, "unauthorized")
    } else if apperr.IsNotFoundError(err) {
      writeError(w, http.StatusNotFound, "expense not found")
    } else if apperr.IsInvalidError(err) {
      writeError(w, http.StatusBadRequest, "invalid expense ID")
    } else {
      writeError(w, http.StatusInternalServerError, "internal server error")
    }
    return
  }
  expenseDTO := dtos.ToExpenseDTO(expense)
  writeResponse(w, http.StatusOK, "expense fetched successfully", expenseDTO)
}

func parseExpensesFilterOptions(r *http.Request) (domain.ExpensesFilterOptions, error) {
	var err error
	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	page := 0
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	userID := r.URL.Query().Get("user_id")

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return domain.ExpensesFilterOptions{}, err
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return domain.ExpensesFilterOptions{}, err
		}
	}

	return domain.ExpensesFilterOptions{
		UserID: userID,
		Page:   page,
		Limit:  limit,
		Status: domain.RequestStatus(status),
	}, nil
}
