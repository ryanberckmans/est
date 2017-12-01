package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringsShortForm(t *testing.T) {
	assert.Panics(t, func() {
		stringsShortForm(5, 10, []string{})
	}, "empty slice panics")

	sf := func(s string) []string {
		return strings.Fields(s)
	}
	testCases := []struct {
		name        string
		minTokenLen int
		maxLen      int
		ss          []string
		expected    string
	}{
		{"one long token", 5, 10, sf("abcdefghijklmnopqrstuvwxyz"), "abcdefghij"},
		{"one char", 5, 10, sf("a"), "a"},
		{"one short token below minTokenLen", 5, 10, sf("abc"), "abc"},
		{"one short token above minTokenLen", 5, 10, sf("abcdef"), "abcdef"},
		{"short, long tokens", 5, 8, sf("a bcdefghijklmnopqrstuvwxyz"), "a bcdefg"},
		{"short, long tokens #2", 5, 8, sf("abc defghijklmnopqrstuvwxyz"), "abc defg"},
		{"long, short tokens", 5, 8, sf("abcdefghijklmnopqrstuvwxy z"), "abcdef z"},
		{"long, short tokens #2", 5, 8, sf("abcdefghijklmnopqrstuvw xyz"), "abcde xy"},
		{"long, short tokens #3", 1, 8, sf("abcdefghijklmnopqrstuvw xyz"), "abcd xyz"},
		{"short, short, short tokens", 5, 20, sf("a b c"), "a b c"},
		{"short, short, short tokens #2", 5, 20, sf("ab cd ef"), "ab cd ef"},
		{"short, short, long tokens", 1, 8, sf("a b cdefghijklmnopqrstuvw"), "a b cdef"},
		{"long, long, long tokens", 6, 15, sf("abcdefghijklmnopqrstuvwxyz 2abcdefghijklmnopqrstuvwxyz 3abcdefghijklmnopqrstuvwxyz"), "abcdef 2abcde 3"},
		{"long, long, long tokens #2", 6, 17, sf("abcdefghijklmnopqrstuvwxyz 2abcdefghijklmnopqrstuvwxyz 3abcdefghijklmnopqrstuvwxyz"), "abcdef 2abcde 3ab"},
		{"no room for tail", 1, 2, sf("a b cdefghijklmnopqrstuvw"), "a"},
		{"no room for tail #2", 3, 4, sf("abc def ghi"), "abc"},
		{"long tokens", 3, 10, sf("manipulate experiments"), "man experi"}, // a weakness of this algorithm is that it favors the tail when head/tail exceed minTokenLength
	}
	testsRun := 0
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testsRun++
			assert.Equal(t, tc.expected, stringsShortForm(tc.minTokenLen, tc.maxLen, tc.ss), "stringsShortForm failed for %s", tc.name)
		})
	}
	assert.Equal(t, len(testCases), testsRun, "sanity check that all tests ran")
}
