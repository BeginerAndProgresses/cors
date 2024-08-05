package utils

import "net/http"

// GetHeaderFirst 获取header中key第一个的值
func GetHeaderFirst(hdr http.Header, key string) ([]string, bool) {
	v, ok := hdr[key]
	if !ok || len(v) == 0 {
		return nil, false
	}
	return v[:1], true
}
