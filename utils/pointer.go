package utils

import "time"

func IntPointer(i int) *int {
	return &i
}

func Int64Pointer(i int64) *int64 {
	return &i
}

func Float64Pointer(f float64) *float64 {
	return &f
}

func TimePointer(t time.Time) *time.Time {
	return &t
}

func StringPointer(s string) *string {
	return &s
}

func BoolPointer(b bool) *bool {
	return &b
}
