package core

import (
	"fmt"
	"os"
	"strings"
)

// WithEstConfigAndFile is the standard entrypoint into est/core.
// Loads or creates a canonical estconfig and estfile, then passes
// them to the passed function.
// TODO drop the With() and should just be (es, ef, error)
// TODO ensure all fatal/errors in entire app written to stderr
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

	ef2 := toExportedEstfile(ef)
	fn(&ec, &ef2)
}
