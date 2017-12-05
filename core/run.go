package core

import (
	"fmt"
	"os"
	"strings"
)

// WithEstConfigAndFile is the standard entrypoint into est/core.
// Loads or creates a canonical estconfig and estfile, then passes
// them to the passed function.
func WithEstConfigAndFile(fn func(ec *EstConfig, ef *EstFile), failFn func()) {
	ec, err := getEstConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s", err)
		failFn()
		return
	}

	estFileName := strings.Replace(ec.Estfile, "$HOME", os.Getenv("HOME"), -1)
	ef, err := getEstFile(estFileName) // TODO support replacement of any env
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s", err)
		failFn()
		return
	}
	ef.fileName = estFileName

	fn(&ec, &ef)
}
