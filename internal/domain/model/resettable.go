package model

// generate:reset
type ResetableStruct struct {
	IntVal   int
	StrVal   string
	StrPtr   *string
	IntSlice []int
	StrMap   map[string]string
	Child    *ResetableStruct
}
