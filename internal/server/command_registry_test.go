package server

import (
	"testing"

	"github.com/tidwall/redcon"
)

func TestCommandRegistry_GetCommand(t *testing.T) {
	registry := NewCommandRegistry()

	tests := []struct {
		name        string
		commandName string
		shouldExist bool
		expectedCmd string
	}{
		{"existing command", "SET", true, "SET"},
		{"existing command lowercase", "set", true, "SET"},
		{"alias command", "DELETE", true, "DEL"},
		{"non-existing command", "NONEXISTENT", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, exists := registry.GetCommand(tt.commandName)

			if exists != tt.shouldExist {
				t.Errorf("GetCommand(%s) exists = %v, want %v", tt.commandName, exists, tt.shouldExist)
			}

			if exists && cmd.Name != tt.expectedCmd {
				t.Errorf("GetCommand(%s) name = %s, want %s", tt.commandName, cmd.Name, tt.expectedCmd)
			}
		})
	}
}

func TestCommandRegistry_ValidateCommand(t *testing.T) {
	registry := NewCommandRegistry()

	tests := []struct {
		name      string
		command   string
		args      [][]byte
		shouldErr bool
	}{
		{"SET valid args", "SET", [][]byte{[]byte("SET"), []byte("key"), []byte("value")}, false},
		{"SET too few args", "SET", [][]byte{[]byte("SET"), []byte("key")}, true},
		{"SET too many args", "SET", [][]byte{[]byte("SET"), []byte("key"), []byte("value"), []byte("extra")}, true},
		{"GET valid args", "GET", [][]byte{[]byte("GET"), []byte("key")}, false},
		{"GET too few args", "GET", [][]byte{[]byte("GET")}, true},
		{"DEL valid multiple args", "DEL", [][]byte{[]byte("DEL"), []byte("key1"), []byte("key2")}, false},
		{"DEL too few args", "DEL", [][]byte{[]byte("DEL")}, true},
		{"PING valid single arg", "PING", [][]byte{[]byte("PING")}, false},
		{"PING valid two args", "PING", [][]byte{[]byte("PING"), []byte("message")}, false},
		{"PING too many args", "PING", [][]byte{[]byte("PING"), []byte("msg1"), []byte("msg2")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, exists := registry.GetCommand(tt.command)
			if !exists {
				t.Fatalf("Command %s not found in registry", tt.command)
			}

			cmd := redcon.Command{Args: tt.args}
			err := registry.ValidateCommand(cmd, metadata)

			hasErr := err != nil
			if hasErr != tt.shouldErr {
				t.Errorf("ValidateCommand(%s) error = %v, want error = %v", tt.command, hasErr, tt.shouldErr)
			}
		})
	}
}

func TestCommandRegistry_Aliases(t *testing.T) {
	registry := NewCommandRegistry()

	cmd, exists := registry.GetCommand("DELETE")
	if !exists {
		t.Fatal("DELETE alias should exist")
	}

	if cmd.Name != "DEL" {
		t.Errorf("DELETE alias should resolve to DEL, got %s", cmd.Name)
	}
}
