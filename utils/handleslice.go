package utils

import "slices"

func HandleSlice[T any](slice []T, fn func(T) T) []T {
	outSlice := make([]T, len(slice))
	for i, v := range slice {
		outSlice[i] = fn(v)
	}
	return outSlice
}

// ContainerOtherAll 一个切片是否包含另一个切片所有元素
func ContainerOtherAll[T comparable](so, st []T) bool {
	for _, t := range st {
		if !slices.Contains(so, t) {
			return false
		}
	}
	return true
}
