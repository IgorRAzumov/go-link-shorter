package model

import "errors"

var ErrEmptyUser = errors.New("empty userID")

var ErrURLConflict = errors.New("URL already exists")

var ErrShortKeyCreation = errors.New("short key creation error")

var ErrURLDeleted = errors.New("URL is deleted")

type URLConflictError struct {
	ExistingShortKey string
}

func (e *URLConflictError) Error() string {
	return ErrURLConflict.Error()
}
