package main

import (
	"os"

	"github.com/ryanberckmans/est/core"
)

func main() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.NotDeleted().Started().SortByStartedAtDescending()
		os.Stdout.WriteString(renderPrompt(ts))
	}, func() {
		// failed to load estconfig or estfile
		os.Stdout.WriteString(promptFailed)
	})
}
