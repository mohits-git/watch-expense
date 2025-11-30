package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockImageUploadService is a mock implementation of ports.ImageUploadService
type MockImageUploadService struct {
	mock.Mock
}

func (m *MockImageUploadService) UploadImage(ctx context.Context, body io.Reader) (string, error) {
	args := m.Called(ctx, body)
	return args.String(0), args.Error(1)
}

func (m *MockImageUploadService) DeleteImage(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func Test_handlers_ImageUploadHandler_HandleUploadImage(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*MockImageUploadService)
		expectedStatus int
		expectedMsg    string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful upload",
			requestBody: "image data",
			setupMock: func(m *MockImageUploadService) {
				m.On("UploadImage", mock.Anything, mock.Anything).
					Return("/images/test-image-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "Successfully uploaded image.",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Status  int
					Message string
					Data    dtos.ImageUploadResponse
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "/images/test-image-123", response.Data.ImageURL)
			},
		},
		{
			name:        "upload error",
			requestBody: "image data",
			setupMock: func(m *MockImageUploadService) {
				m.On("UploadImage", mock.Anything, mock.Anything).
					Return("", errors.New("upload failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Something went wrong while uploading your image.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockImageUploadService)
			tt.setupMock(mockService)

			handler := NewImageUploadHandler(mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/images", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			handler.HandleUploadImage(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			if tt.checkResponse != nil {
				req = httptest.NewRequest(http.MethodPost, "/api/images", strings.NewReader(tt.requestBody))
				w = httptest.NewRecorder()
				handler.HandleUploadImage(w, req)
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ImageUploadHandler_HandleDeleteImage(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*MockImageUploadService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful delete",
			requestBody: dtos.DeleteImageRequest{
				ImageURL: "/images/test-image-123",
			},
			setupMock: func(m *MockImageUploadService) {
				m.On("DeleteImage", mock.Anything, "/images/test-image-123").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Successfully deleted image.",
		},
		{
			name: "delete error",
			requestBody: dtos.DeleteImageRequest{
				ImageURL: "/images/test-image-123",
			},
			setupMock: func(m *MockImageUploadService) {
				m.On("DeleteImage", mock.Anything, "/images/test-image-123").
					Return(errors.New("delete failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Could not delete the image",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"image_url": 124,
			},
			setupMock:      func(m *MockImageUploadService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockImageUploadService)
			tt.setupMock(mockService)

			handler := NewImageUploadHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodDelete, "/api/images", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleDeleteImage(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}
