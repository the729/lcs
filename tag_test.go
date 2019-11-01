package lcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {
	tag := "enum=test,option,len=16"
	m := parseTag(tag)
	assert.Equal(t, m, map[string]string{
		"enum":   "test",
		"option": "",
		"len":    "16",
	})
}
