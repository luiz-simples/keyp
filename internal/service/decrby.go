package service

func (handler *Handler) decrby(args Args) *Result {
	return processIntegerModification(args, handler.storage.DecrBy, handler)
}
