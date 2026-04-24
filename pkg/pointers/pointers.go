package pointers

import "time"

func NewBoolPointer(b bool) *bool {
	return &b
}

func NewIntPointer(i int) *int {
	return &i
}

func NewUintPointer(u uint) *uint {
	return &u
}

func NewTimePointer(t time.Time) *time.Time{
	return &t
}


func NewStringPointer(s string) *string {
	return &s
}