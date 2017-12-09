package cmd

import (
	"os"
	"strings"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	Long:  `List tasks, TODO better help msg`,
	Run: func(cmd *cobra.Command, args []string) {
		doLS()
	},
}

// TODO ls to allow filtering by status. Thought is to show started and unestimated tasks by default. Also have an --all option to simplify showing everything.

func doLS() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.SortByStatusDescending()
		rs := make([]string, len(ts)+1) // +1 causes the last element to be empty string, which causes the Join to add an extra newline
		for i := range ts {
			rs[i] = core.RenderTaskOneLineSummary(ts[i])
		}
		os.Stdout.WriteString(strings.Join(rs, "\n"))
	}, func() {
		// failed to load estconfig or estfile. Err printed elsewhere.
		os.Exit(1)
	})
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
