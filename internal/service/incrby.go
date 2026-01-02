package service

func (handler *Handler) incrby(args Args) *Result {
	return processIntegerModification(args, handler.storage.IncrBy, handler)
}
