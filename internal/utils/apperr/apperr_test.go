package apperr

import (
	"errors"
	"testing"
)

func Test_apperr_AppError_Error(t *testing.T) {
	type fields struct {
		Code    AppErrorCode
		Message string
		Err     error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "without wrapped error",
			fields: fields{
				Code:    ErrNotFound,
				Message: "resource not found",
				Err:     nil,
			},
			want: "resource not found",
		},
		{
			name: "with wrapped error",
			fields: fields{
				Code:    ErrInternal,
				Message: "internal server error",
				Err:     errors.New("database connection failed"),
			},
			want: "internal server error: database connection failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AppError{
				Code:    tt.fields.Code,
				Message: tt.fields.Message,
				Err:     tt.fields.Err,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_NewAppError(t *testing.T) {
	type args struct {
		code    AppErrorCode
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		{
			name: "create new AppError without wrapped error",
			args: args{
				code:    ErrInvalid,
				message: "invalid input",
				err:     nil,
			},
			want: &AppError{
				Code:    ErrInvalid,
				Message: "invalid input",
				Err:     nil,
			},
		},
		{
			name: "create new AppError with wrapped error",
			args: args{
				code:    ErrInternal,
				message: "internal error",
				err:     errors.New("some internal error"),
			},
			want: &AppError{
				Code:    ErrInternal,
				Message: "internal error",
				Err:     errors.New("some internal error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAppError(tt.args.code, tt.args.message, tt.args.err); !errors.Is(got, tt.want) && got.Code != tt.want.Code && got.Message != tt.want.Message {
				t.Errorf("NewAppError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrNotFound code",
			err: &AppError{
				Code:    ErrNotFound,
				Message: "resource not found",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrInternal,
				Message: "internal error",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrUnauthorized code",
			err: &AppError{
				Code:    ErrUnauthorized,
				Message: "unauthorized access",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrForbidden,
				Message: "forbidden access",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorizedError(tt.err); got != tt.want {
				t.Errorf("IsUnauthorizedError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsForbiddenError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrForbidden code",
			err: &AppError{
				Code:    ErrForbidden,
				Message: "forbidden access",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrInvalid,
				Message: "invalid input",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbiddenError(tt.err); got != tt.want {
				t.Errorf("IsForbiddenError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsConflictError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrConflict code",
			err: &AppError{
				Code:    ErrConflict,
				Message: "resource conflict",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrTimeout,
				Message: "operation timed out",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConflictError(tt.err); got != tt.want {
				t.Errorf("IsConflictError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsInvalidError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrInvalid code",
			err: &AppError{
				Code:    ErrInvalid,
				Message: "invalid input",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrNotFound,
				Message: "resource not found",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidError(tt.err); got != tt.want {
				t.Errorf("IsInvalidError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsInternalError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrInternal code",
			err: &AppError{
				Code:    ErrInternal,
				Message: "internal server error",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrTimeout,
				Message: "operation timed out",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInternalError(tt.err); got != tt.want {
				t.Errorf("IsInternalError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apperr_IsTimeoutError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-AppError type",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "AppError with ErrTimeout code",
			err: &AppError{
				Code:    ErrTimeout,
				Message: "operation timed out",
				Err:     nil,
			},
			want: true,
		},
		{
			name: "AppError with different code",
			err: &AppError{
				Code:    ErrConflict,
				Message: "resource conflict",
				Err:     nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeoutError(tt.err); got != tt.want {
				t.Errorf("IsTimeoutError() = %v, want %v", got, tt.want)
			}
		})
	}
}
