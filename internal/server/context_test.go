package server

import (
	"context"
	"testing"
	"time"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

func TestContextCancellation(t *testing.T) {
	registry := NewCommandRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	cmd := redcon.Command{
		Args: [][]byte{
			[]byte("SET"),
			[]byte("testkey"),
			[]byte("testvalue"),
		},
	}

	metadata, exists := registry.GetCommand("SET")
	if !exists {
		t.Fatal("SET command should exist in registry")
	}

	err := registry.ValidateCommand(cmd, metadata)
	if err != nil {
		t.Fatalf("Command validation failed: %v", err)
	}

	if ctx.Err() == nil {
		t.Error("Context should be canceled by now")
	}

	if !isContextCanceled(ctx.Err()) {
		t.Error("isContextCanceled should return true for canceled context")
	}
}

func TestContextValidation(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		shouldMatch bool
	}{
		{"context canceled", context.Canceled, true},
		{"context deadline exceeded", context.DeadlineExceeded, true},
		{"other error", storage.ErrKeyNotFound, false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isContextCanceled(tt.err)
			if result != tt.shouldMatch {
				t.Errorf("isContextCanceled(%v) = %v, want %v", tt.err, result, tt.shouldMatch)
			}
		})
	}
}
