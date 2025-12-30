package domain

import "errors"

var OK []byte = []byte("OK")
var CANCELED error = errors.New("ERR operation canceled")

func NewResult() *Result {
	result := Result{}
	return result.Clear()
}

func (result *Result) SetCanceled() *Result {
	result.Error = CANCELED
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
