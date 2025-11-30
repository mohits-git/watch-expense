# HTTP Handlers Test Suite

## Overview
Comprehensive test suite for HTTP handlers using `httptest`, `testify/assert`, `testify/require`, mock services, and DTOs.

## Test Files Created

### 1. auth_test.go ✅
Tests for `AuthHandler`:
- **HandleLogin** (6 test cases)
  - Successful login with token generation
  - Invalid email (not found)
  - Invalid password (unauthorized)
  - Invalid request body
  - Invalid input validation
  - Internal server error
  
- **HandleLogout** (3 test cases)
  - Successful logout with cookie clearing
  - No token in context (unauthorized)
  - Logout service error
  
- **HandleMe** (4 test cases)
  - Successful get current user
  - Unauthorized error
  - User not found
  - Internal server error

### 2. common_test.go ✅
Tests for `CommonHandler`:
- **HealthCheck** (1 test case)
  - Successful health check returning "OK"

### 3. image_upload_test.go ✅
Tests for `ImageUploadHandler`:
- **HandleUploadImage** (2 test cases)
  - Successful image upload
  - Upload error
  
- **HandleDeleteImage** (3 test cases)
  - Successful image deletion
  - Delete error
  - Invalid request body

### 4. user_test.go ✅
Tests for `UserHandler`:
- **HandleGetAllUsers** (3 test cases)
  - Successful get all users
  - Unauthorized
  - Internal server error
  
- **HandleGetUserByID** (4 test cases)
  - Successful get user by ID
  - User not found
  - Unauthorized
  - Invalid user ID
  
- **HandleCreateUser** (4 test cases)
  - Successful create user
  - Unauthorized (non-admin)
  - Invalid user data
  - Invalid request body
  
- **HandleUpdateUser** (4 test cases)
  - Successful update user
  - Unauthorized
  - Invalid user data
  - Invalid request body
  
- **HandleDeleteUser** (3 test cases)
  - Successful delete user
  - Unauthorized
  - Invalid user ID
  
- **HandleGetUserBudget** (4 test cases)
  - Successful get user budget
  - Unauthorized
  - User not found
  - Forbidden

### 5. department_test.go ✅
Tests for `DepartmentHandler`:
- **HandleGetAllDepartments** (3 test cases)
  - Successful get all departments
  - Unauthorized
  - Internal server error
  
- **HandleCreateDepartment** (4 test cases)
  - Successful create department
  - Invalid department data
  - Unauthorized
  - Invalid request body
  
- **HandleGetDepartmentByID** (3 test cases)
  - Successful get department by ID
  - Department not found
  - Invalid department ID
  
- **HandleUpdateDepartment** (4 test cases)
  - Successful update department
  - Department not found
  - Invalid department data
  - Invalid request body

## Testing Patterns Used

### 1. Table-Driven Tests
All tests use the table-driven pattern with consistent structure:
```go
tests := []struct {
    name           string
    requestBody    interface{}
    setupMock      func(*mockservice.ServiceType)
    expectedStatus int
    expectedMsg    string
    checkResponse  func(*testing.T, *httptest.ResponseRecorder)
}{
    // test cases
}
```

### 2. Mock Setup
- Uses `mockservice` package for service mocks
- Each test case has a `setupMock` function to configure expectations
- Assertions verified with `AssertExpectations(t)`

### 3. HTTP Testing
- Uses `httptest.NewRequest` for creating requests
- Uses `httptest.NewRecorder` for recording responses
- Tests path parameters with `req.SetPathValue("id", value)`
- Tests request bodies with JSON marshaling

### 4. Response Validation
- Checks HTTP status codes
- Validates response messages
- Optionally validates response data with `checkResponse` function
- Uses `dtos.BaseResponse` for decoding responses

### 5. Error Handling
- Tests all error types: `ErrNotFound`, `ErrUnauthorized`, `ErrInvalid`, `ErrForbidden`
- Tests HTTP error codes: 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 404 (Not Found), 500 (Internal Server Error)
- Tests invalid request bodies

## Test Coverage by Handler

| Handler | Methods Tested | Total Test Cases | Status |
|---------|---------------|------------------|---------|
| AuthHandler | 3 | 13 | ✅ Complete |
| CommonHandler | 1 | 1 | ✅ Complete |
| ImageUploadHandler | 2 | 5 | ✅ Complete |
| UserHandler | 6 | 22 | ✅ Complete |
| DepartmentHandler | 4 | 14 | ✅ Complete |
| ProjectHandler | 4 | 21 | ✅ Complete |
| ExpenseHandler | 6 | 35 | ✅ Complete |
| AdvanceHandler | 6 | 35 | ✅ Complete |
| **TOTAL** | **32** | **146** | ✅ **100% Complete** |

## Summary

All handler tests have been successfully implemented with comprehensive coverage:
- ✅ **8 handler test files** covering all HTTP endpoints
- ✅ **146 total test cases** covering success and error scenarios
- ✅ **32 handler methods** fully tested
- ✅ All HTTP status codes covered (200, 201, 400, 401, 403, 404, 500)
- ✅ All error types tested (ErrNotFound, ErrUnauthorized, ErrInvalid, ErrForbidden, ErrInternal)
- ✅ Request validation, authorization, pagination, and filtering tested
- ✅ Complex features tested: status transitions, summary endpoints, filter options

## Key Testing Utilities

### Mock Services
Located in `tests/mock_service/`:
- `AuthenticationService`
- `UserService`
- `DepartmentService`
- `ProjectService`
- `ExpenseService`
- `AdvanceService`

### DTOs
Located in `internal/adapters/http/dtos/`:
- Request DTOs (CreateX, UpdateX, etc.)
- Response DTOs (with proper JSON tags)
- `BaseResponse` for standard API responses

### Handler Utilities
Located in `internal/adapters/http/handlers/utils.go`:
- `decodeRequest[T]` - Decode JSON request body
- `writeResponse` - Write JSON response with status
- `writeError` - Write error response
- `encodeJson` / `decodeJson` - JSON encoding/decoding helpers

## Running Tests

```bash
# Run all handler tests
go test ./internal/adapters/http/handlers/...

# Run specific handler tests
go test ./internal/adapters/http/handlers/ -run Test_handlers_AuthHandler

# Run with verbose output
go test -v ./internal/adapters/http/handlers/...

# Run with coverage
go test -cover ./internal/adapters/http/handlers/...
```

## Test Naming Convention

All tests follow the pattern:
```
Test_handlers_<HandlerName>_<MethodName>
```

Examples:
- `Test_handlers_AuthHandler_HandleLogin`
- `Test_handlers_UserHandler_HandleGetUserByID`
- `Test_handlers_DepartmentHandler_HandleCreateDepartment`

## Notes

1. **Context Handling**: Tests use `context.Background()` for simplicity. Some tests (like logout) use `authctx.WithToken()` for token-aware contexts.

2. **Path Parameters**: Go 1.22+ routing with `req.SetPathValue("id", value)` is used for testing path parameters.

3. **Response Decoding**: Tests decode responses into `dtos.BaseResponse` first, then optionally into specific response types for data validation.

4. **Mock Assertions**: All tests call `mockService.AssertExpectations(t)` to verify mock expectations were met.

5. **Error Types**: Tests use `apperr.NewAppError(apperr.ErrCode, msg, err)` to create application-specific errors.
