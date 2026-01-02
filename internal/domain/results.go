package domain

func NewResult() *Result {
	result := Result{}
	return result.Clear()
}

func (result *Result) SetCanceled() *Result {
	result.Error = ErrCanceled
	result.Response = nil
	return result
}

func (result *Result) Clear() *Result {
	result.Error = nil
	result.Response = nil
	return result
}

func (result *Result) SetOK() *Result {
	result.Response = OK
	result.Error = nil
	return result
}
