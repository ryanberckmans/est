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
		ts := ef.Tasks.SortByStatusDescending()
		if !lsFlagDeleted {
			ts = ts.IsNotDeleted()
		}
		if !lsFlagDone {
			ts = ts.IsNotDone()
		}
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

var lsFlagDone bool    // show done tasks
var lsFlagDeleted bool // show deleted tasks

func init() {
	lsCmd.PersistentFlags().BoolVarP(&lsFlagDone, "done", "d", false, "show done tasks")
	lsCmd.PersistentFlags().BoolVarP(&lsFlagDeleted, "deleted", "D", false, "show deleted tasks")
	rootCmd.AddCommand(lsCmd)
}
