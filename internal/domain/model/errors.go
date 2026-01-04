package model

import "errors"

var ErrURLConflict = errors.New("URL already exists")

type URLConflictError struct {
	ExistingShortKey string
}

func (e *URLConflictError) Error() string {
	return ErrURLConflict.Error()
}
