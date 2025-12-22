package server

import (
	"context"

	"github.com/tidwall/redcon"
)

type CommandHandler func(context.Context, redcon.Conn, redcon.Command)

type CommandMetadata struct {
	Name    string
	MinArgs int
	MaxArgs int
	Handler CommandHandler
	Aliases []string
}

type CommandRegistry struct {
	commands map[string]*CommandMetadata
	aliases  map[string]string
}

func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]*CommandMetadata),
		aliases:  make(map[string]string),
	}

	registry.registerCommands()
	return registry
}

func (registry *CommandRegistry) registerCommands() {
	commands := []*CommandMetadata{
		{
			Name:    "PING",
			MinArgs: 1,
			MaxArgs: 2,
			Handler: wrapHandlerPing,
			Aliases: []string{},
		},
		{
			Name:    "SET",
			MinArgs: 3,
			MaxArgs: 3,
			Handler: nil,
			Aliases: []string{},
		},
		{
			Name:    "GET",
			MinArgs: 2,
			MaxArgs: 2,
			Handler: nil,
			Aliases: []string{},
		},
		{
			Name:    "DEL",
			MinArgs: 2,
			MaxArgs: -1,
			Handler: nil,
			Aliases: []string{"DELETE"},
		},
		{
			Name:    "EXPIRE",
			MinArgs: 3,
			MaxArgs: 3,
			Handler: nil,
			Aliases: []string{},
		},
		{
			Name:    "EXPIREAT",
			MinArgs: 3,
			MaxArgs: 3,
			Handler: nil,
			Aliases: []string{},
		},
		{
			Name:    "TTL",
			MinArgs: 2,
			MaxArgs: 2,
			Handler: nil,
			Aliases: []string{},
		},
		{
			Name:    "PTTL",
			MinArgs: 2,
			MaxArgs: 2,
			Handler: nil,
			Aliases: []string{},
		},
		{
			Name:    "PERSIST",
			MinArgs: 2,
			MaxArgs: 2,
			Handler: nil,
			Aliases: []string{},
		},
	}

	for _, cmd := range commands {
		registry.registerCommand(cmd)
	}
}

func (registry *CommandRegistry) registerCommand(metadata *CommandMetadata) {
	registry.commands[metadata.Name] = metadata

	for _, alias := range metadata.Aliases {
		registry.aliases[alias] = metadata.Name
	}
}

func (registry *CommandRegistry) GetCommand(name string) (*CommandMetadata, bool) {
	name = normalizeCommandName(name)

	if realName, isAlias := registry.aliases[name]; isAlias {
		name = realName
	}

	cmd, exists := registry.commands[name]
	return cmd, exists
}

func (registry *CommandRegistry) ValidateCommand(cmd redcon.Command, metadata *CommandMetadata) error {
	argCount := len(cmd.Args)

	if argCount < metadata.MinArgs {
		return newInvalidArgsError(metadata.Name)
	}

	if metadata.MaxArgs > 0 && argCount > metadata.MaxArgs {
		return newInvalidArgsError(metadata.Name)
	}

	return nil
}

func wrapHandlerPing(ctx context.Context, conn redcon.Conn, cmd redcon.Command) {
	handlePing(conn, cmd)
}
