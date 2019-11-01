package lcs

import (
	"strings"
)

func parseTag(tag string) map[string]string {
	m := make(map[string]string)
	for _, option := range strings.Split(tag, ",") {
		var key, value string
		kv := strings.SplitN(option, "=", 2)
		if len(kv) > 0 {
			key = strings.TrimSpace(kv[0])
		}
		if len(kv) > 1 {
			value = strings.TrimSpace(kv[1])
		}
		m[key] = value
	}
	return m
}
