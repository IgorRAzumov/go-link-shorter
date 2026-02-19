package model

import "errors"

// ErrEmptyUser возвращается при пустом userID.
var ErrEmptyUser = errors.New("empty userID")

// ErrURLConflict возвращается при попытке сохранить уже существующий URL.
var ErrURLConflict = errors.New("URL already exists")

// ErrShortKeyCreation возвращается при ошибке генерации короткого ключа.
var ErrShortKeyCreation = errors.New("short key creation error")

// ErrURLDeleted возвращается при обращении к удалённой ссылке.
var ErrURLDeleted = errors.New("URL is deleted")

// URLConflictError возвращается при конфликте URL с указанием существующего short_key.
type URLConflictError struct {
	ExistingShortKey string // Короткий ключ уже существующей ссылки
}

func (e *URLConflictError) Error() string {
	return ErrURLConflict.Error()
}
