package pointers

func NewBoolPointer(b bool) *bool {
	return &b
}

func NewIntPointer(i int) *int {
	return &i
}

func NewUintPointer(u uint) *uint {
	return &u
}
