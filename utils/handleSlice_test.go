package utils

import (
	"strings"
	"testing"
)

func TestHandleSlice(t *testing.T) {
	a := []string{"a", "b", "c"}
	a = HandleSlice(a, strings.ToUpper)
	t.Log(a)
}
