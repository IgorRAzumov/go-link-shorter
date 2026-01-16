package healthcheck

import (
	"context"
	"errors"
	"testing"
)

type MockHealthCheckService struct {
	CheckStorageConnectionFunc func(ctx context.Context) (bool, error)
}

func (mock *MockHealthCheckService) CheckStorageConnection(ctx context.Context) (bool, error) {
	if mock.CheckStorageConnectionFunc != nil {
		return mock.CheckStorageConnectionFunc(ctx)
	}
	return false, nil
}

func TestNewHealthCheckUsecase(t *testing.T) {
	mockService := &MockHealthCheckService{}
	usecase := NewHealthCheckUsecase(mockService)

	if usecase == nil {
		t.Fatal("NewHealthCheckUsecase returned nil")
	}
	if usecase.service == nil {
		t.Error("Usecase.service should not be nil")
	}
}

func TestUsecase_CheckSystemConnections(t *testing.T) {
	mockService := &MockHealthCheckService{
		CheckStorageConnectionFunc: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	}
	usecase := NewHealthCheckUsecase(mockService)

	ctx := context.Background()
	result, err := usecase.CheckSystemConnections(ctx)
	if err != nil {
		t.Fatalf("CheckSystemConnections failed: %v", err)
	}
	if !result {
		t.Error("CheckSystemConnections should return true for successful connection")
	}
}

func TestUsecase_CheckSystemConnections_ReturnsFalse(t *testing.T) {
	mockService := &MockHealthCheckService{
		CheckStorageConnectionFunc: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	}
	usecase := NewHealthCheckUsecase(mockService)

	ctx := context.Background()
	result, err := usecase.CheckSystemConnections(ctx)
	if err != nil {
		t.Errorf("CheckSystemConnections should not return error, got: %v", err)
	}
	if result {
		t.Error("CheckSystemConnections should return false for failed connection")
	}
}

func TestUsecase_CheckSystemConnections_WithError(t *testing.T) {
	expectedError := errors.New("connection error")
	mockService := &MockHealthCheckService{
		CheckStorageConnectionFunc: func(ctx context.Context) (bool, error) {
			return false, expectedError
		},
	}
	usecase := NewHealthCheckUsecase(mockService)

	ctx := context.Background()
	result, err := usecase.CheckSystemConnections(ctx)
	if err == nil {
		t.Fatal("CheckSystemConnections should return error")
	}
	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
	if result {
		t.Error("CheckSystemConnections should return false when error occurs")
	}
}

type testContextKey string

func TestUsecase_CheckSystemConnections_PassesContext(t *testing.T) {
	testKey := testContextKey("test-key")
	testValue := "test-value"
	ctx := context.WithValue(context.Background(), testKey, testValue)
	contextPassed := false

	mockService := &MockHealthCheckService{
		CheckStorageConnectionFunc: func(passedContext context.Context) (bool, error) {
			if passedContext.Value(testKey) == testValue {
				contextPassed = true
			}
			return true, nil
		},
	}
	usecase := NewHealthCheckUsecase(mockService)

	_, err := usecase.CheckSystemConnections(ctx)
	if err != nil {
		t.Fatalf("CheckSystemConnections failed: %v", err)
	}
	if !contextPassed {
		t.Error("Context should be passed to service")
	}
}
