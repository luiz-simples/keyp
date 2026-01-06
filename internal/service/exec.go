package service

func (handler *Handler) exec() Results {
	results := make(Results, 0, len(handler.multArgs))

	if handler.multEnabled {
		for _, args := range handler.multArgs {
			results = append(results, handler.do(args))
		}
	}

	return results
}
