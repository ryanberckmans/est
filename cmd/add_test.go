package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLooksLikeIdPrefix(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected bool
	}{
		{"prefix starts number", "5d", true},
		{"prefix ends number", "a8", true},
		{"prefix bigger", "3c6a8f", true},
		{"two tokens", "3c6a8f a8", false},
		{"empty", "", false},
		{"no numbers", "hello", false},
		{"chars outside hexadecimal range", "5q", false},
		{"no numbers, all inside hex range", "abc", false},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			g
			assert.Equal(t, tc.expected, looksLikeIDPrefix(tc.input))
		})
	}
}
