package healthcheck

import (
	"context"
	"errors"
	"testing"
)

type MockHealthCheckRepository struct {
	CheckStorageConnectionFunc func(ctx context.Context) (bool, error)
}

func (mock *MockHealthCheckRepository) CheckStorageConnection(ctx context.Context) (bool, error) {
	if mock.CheckStorageConnectionFunc != nil {
		return mock.CheckStorageConnectionFunc(ctx)
	}
	return false, nil
}

func TestNewHealthCheckService(t *testing.T) {
	mockRepository := &MockHealthCheckRepository{}
	service := NewHealthCheckService(mockRepository)

	if service == nil {
		t.Fatal("NewHealthCheckService returned nil")
	}
	if service.repository == nil {
		t.Error("Service.repository should not be nil")
	}
}

func TestService_CheckStorageConnection(t *testing.T) {
	mockRepository := &MockHealthCheckRepository{
		CheckStorageConnectionFunc: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	}
	service := NewHealthCheckService(mockRepository)

	ctx := context.Background()
	result, err := service.CheckStorageConnection(ctx)
	if err != nil {
		t.Fatalf("CheckStorageConnection failed: %v", err)
	}
	if !result {
		t.Error("CheckStorageConnection should return true for successful connection")
	}
}

func TestService_CheckStorageConnection_ReturnsFalse(t *testing.T) {
	mockRepository := &MockHealthCheckRepository{
		CheckStorageConnectionFunc: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	}
	service := NewHealthCheckService(mockRepository)

	ctx := context.Background()
	result, err := service.CheckStorageConnection(ctx)
	if err != nil {
		t.Errorf("CheckStorageConnection should not return error, got: %v", err)
	}
	if result {
		t.Error("CheckStorageConnection should return false for failed connection")
	}
}

func TestService_CheckStorageConnection_WithError(t *testing.T) {
	expectedError := errors.New("repository error")
	mockRepository := &MockHealthCheckRepository{
		CheckStorageConnectionFunc: func(ctx context.Context) (bool, error) {
			return false, expectedError
		},
	}
	service := NewHealthCheckService(mockRepository)

	ctx := context.Background()
	result, err := service.CheckStorageConnection(ctx)
	if err == nil {
		t.Fatal("CheckStorageConnection should return error")
	}
	if err != expectedError {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
	if result {
		t.Error("CheckStorageConnection should return false when error occurs")
	}
}

type testContextKey string

func TestService_CheckStorageConnection_PassesContext(t *testing.T) {
	testKey := testContextKey("test-key")
	testValue := "test-value"
	ctx := context.WithValue(context.Background(), testKey, testValue)
	contextPassed := false

	mockRepository := &MockHealthCheckRepository{
		CheckStorageConnectionFunc: func(passedContext context.Context) (bool, error) {
			if passedContext.Value(testKey) == testValue {
				contextPassed = true
			}
			return true, nil
		},
	}
	service := NewHealthCheckService(mockRepository)

	_, err := service.CheckStorageConnection(ctx)
	if err != nil {
		t.Fatalf("CheckStorageConnection failed: %v", err)
	}
	if !contextPassed {
		t.Error("Context should be passed to repository")
	}
}
