package model

// generate:reset
type ResettableStruct struct {
	IntVal   int
	StrVal   string
	StrPtr   *string
	IntSlice []int
	StrMap   map[string]string
	Child    *ResettableStruct
}
