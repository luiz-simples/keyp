package service

func (handler *Handler) Clear() {
	handler.multEnabled = false
	handler.multArgs = handler.multArgs[:0]
	handler.context = nil
}
