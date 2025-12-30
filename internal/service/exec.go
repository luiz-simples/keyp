package service

func (handler *Handler) exec() Results {
	results := make(Results, 0, len(handler.multArgs))

	if handler.multEnabled {
		for _, args := range handler.multArgs {
			cmdName := normalizeCommandName(string(args[0]))
			command := handler.commands[cmdName]
			results = append(results, command(args))
		}
	}

	return results
}
