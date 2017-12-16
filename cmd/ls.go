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
	Long:  `List tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		doLS()
	},
}

func doLS() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.SortByStatusDescending().IsNotDeleted()
		rs := make([]string, len(ts)+1) // +1 causes the last element to be empty string, which causes the Join to add an extra newline
		for i := range ts {
			rs[i] = core.RenderTaskOneLineSummary(ts[i], i == 0)
		}
		os.Stdout.WriteString(strings.Join(rs, "\n"))
	}, func() {
		// failed to load estconfig or estfile. Err printed elsewhere.
		os.Exit(1)
	})
}

func init() {
	// TODO --done to show done tasks, default to not showing done
	// TODO --deleted to show deleted tasks, default does not show deleted
	rootCmd.AddCommand(lsCmd)
}
