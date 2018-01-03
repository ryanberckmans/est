package cmd

import (
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var yesterdayCmd = &cobra.Command{
	Use:     "yesterday",
	Aliases: []string{"y"},
	Short:   "Show yesterday's activity",
	Long: `Show yesterday's activity

est yesterday [--ago <duration>]

Show task activity yesterday, where yesterday is defined as the most recent day
with any working hours prior to or including yesterday's calendar date.

Yesterday's calendar date can be further in past with --ago. Note that --ago
uses the same syntax as 'est estimate'; supported units are minutes and hours,
so typically you'll want a multiple of 24 hours.

Examples:
  # Show task activity three days ago
  est y -a48h # this is only 48h, not 72h, because the first 24h is a base

`,
	Run: func(cmd *cobra.Command, args []string) {
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			ts := ef.Tasks.SortByStatusDescending()
			now := applyFlagAgo(time.Now())
			os.Stdout.WriteString(core.RenderYesterdayTasks(globalWorkTimes, ts, now))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

func init() {
	yesterdayCmd.PersistentFlags().StringVarP(&flagAgo, "ago", "a", "", "show activity from one business day prior to today's calendar date minus duration ago")
	rootCmd.AddCommand(yesterdayCmd)
}
