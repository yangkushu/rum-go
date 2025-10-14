package utils

import (
	"golang.org/x/exp/rand"
)

func RandomElement[T any](t []T) T {
	if len(t) == 0 {
		var zeroValue T
		return zeroValue // 返回零值和 false 表示切片为空
	}
	//rand.Seed(uint64(time.Now().UnixNano()))
	index := rand.Intn(len(t))
	return t[index]
}
