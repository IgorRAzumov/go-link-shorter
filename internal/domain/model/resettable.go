package model

// ResettableStruct — структура для тестирования генератора Reset.
// generate:reset
type ResettableStruct struct {
	IntVal   int
	StrVal   string
	StrPtr   *string
	IntSlice []int
	StrMap   map[string]string
	Child    *ResettableStruct
}
