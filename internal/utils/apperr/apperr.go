package apperr

type AppError struct {
	Code    AppErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

type AppErrorCode int

const (
	ErrNone AppErrorCode = iota
	ErrNotFound
	ErrUnauthorized
	ErrForbidden
	ErrConflict
	ErrInvalid
	ErrInternal
	ErrTimeout
)

func NewAppError(code AppErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func IsNotFoundError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrNotFound
}

func IsUnauthorizedError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrUnauthorized
}

func IsForbiddenError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrForbidden
}

func IsConflictError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrConflict
}

func IsInvalidError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrInvalid
}

func IsInternalError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrInternal
}

func IsTimeoutError(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == ErrTimeout
}
