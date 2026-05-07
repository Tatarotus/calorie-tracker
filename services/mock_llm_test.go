package services

import (
	"errors"
	"testing"
)

func TestNewMockLLMProvider(t *testing.T) {
	mock := NewMockLLMProvider()
	if mock == nil {
		t.Fatal("Expected non-nil mock")
	}
}

func TestMockLLMProvider_SetResponse(t *testing.T) {
	mock := NewMockLLMProvider()
	mock.SetResponse("test response")

	if mock.Response != "test response" {
		t.Errorf("Expected response 'test response', got %s", mock.Response)
	}
}

func TestMockLLMProvider_SetError(t *testing.T) {
	mock := NewMockLLMProvider()
	mock.SetError(errors.New("test error"))

	if mock.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestMockLLMProvider_Reset(t *testing.T) {
	mock := NewMockLLMProvider()
	mock.SetResponse("test")
	mock.SetError(errors.New("error"))
	mock.Reset()

	if mock.Response != "" {
		t.Error("Expected response to be reset")
	}
	if mock.Error != nil {
		t.Error("Expected error to be reset")
	}
	if mock.CallCount != 0 {
		t.Error("Expected call count to be reset")
	}
}

func TestMockLLMProvider_Call(t *testing.T) {
	mock := NewMockLLMProvider()

	// Test with fixed response
	mock.SetResponse("hello")
	resp, err := mock.Call("model", "test prompt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp != "hello" {
		t.Errorf("Expected 'hello', got %s", resp)
	}
	if mock.CallCount != 1 {
		t.Errorf("Expected call count 1, got %d", mock.CallCount)
	}
	if mock.LastModel != "model" {
		t.Errorf("Expected model 'model', got %s", mock.LastModel)
	}
	if mock.LastPrompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got %s", mock.LastPrompt)
	}

	// Test with fixed error
	mock.Reset()
	mock.SetError(errors.New("error"))
	_, err = mock.Call("model", "test prompt")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestMockLLMProvider_SetFixedResponse(t *testing.T) {
	mock := NewMockLLMProvider()
	mock.SetFixedResponse("food-model", "food response")
	mock.SetFixedResponse("review-model", "review response")

	resp, err := mock.Call("food-model", "prompt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp != "food response" {
		t.Errorf("Expected 'food response', got %s", resp)
	}

	resp, err = mock.Call("review-model", "prompt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp != "review response" {
		t.Errorf("Expected 'review response', got %s", resp)
	}
}
