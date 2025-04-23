package task

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name       string
		maxWorkers int64
	}{
		{"default workers", 0},
		{"custom workers", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(tt.maxWorkers)
			if server == nil {
				t.Fatal("Server should not be nil")
			}
			if tt.maxWorkers == 0 && server.maxWorkers != 1 {
				t.Errorf("Expected maxWorkers to be 1 when 0 is provided, got %d", server.maxWorkers)
			}
			if tt.maxWorkers > 0 && server.maxWorkers != tt.maxWorkers {
				t.Errorf("Expected maxWorkers to be %d, got %d", tt.maxWorkers, server.maxWorkers)
			}
		})
	}
}

func TestAddTask(t *testing.T) {
	server := New(1)
	defer server.Wait()
	go server.Run()

	taskID, err := server.Add("test data", processorSuccess)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if taskID == uuid.Nil {
		t.Error("Task ID should not be nil")
	}

	task, ok := server.Get(taskID)
	if !ok {
		t.Error("Task should be found in the server")
	}
	if task.Data != "test data" {
		t.Errorf("Expected task data to be 'test data', got '%s'", task.Data)
	}
	if task.Status != Pending {
		t.Errorf("Expected task status to be Pending, got %v", task.Status)
	}
}

func TestProcessing(t *testing.T) {
	server := New(1)
	go server.Run()

	expectedResult := "processed data"
	taskID, err := server.Add(expectedResult, processorSuccess)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	server.Wait()

	task, ok := server.Get(taskID)
	if !ok {
		t.Fatal("Task should be found in the server")
	}

	if task.Status != Done {
		t.Errorf("Expected task status to be Done, got %v", task.Status)
	}
	if task.Result != expectedResult {
		t.Errorf("Expected result to be '%s', got '%s'", expectedResult, task.Result)
	}
}

func TestTaskErrorHandling(t *testing.T) {
	server := New(1)
	go server.Run()

	expectedError := errors.New("ERROR")

	taskID, err := server.Add("test data", processorError)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	server.Wait()

	task, ok := server.Get(taskID)
	if !ok {
		t.Fatal("Task should be found in the server")
	}

	if task.Status != Fail {
		t.Errorf("Expected task status to be Fail, got %v", task.Status)
	}
	if task.Result != expectedError.Error() {
		t.Errorf("Expected error to be '%s', got '%s'", expectedError.Error(), task.Result)
	}
}

func processorSuccess(s string) (string, error) {
	time.Sleep(50 * time.Millisecond)
	return s, nil
}

func processorError(s string) (string, error) {
	time.Sleep(50 * time.Millisecond)
	return "", fmt.Errorf("ERROR")
}
