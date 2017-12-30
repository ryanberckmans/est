package cmd

import (
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var howamidoingCmd = &cobra.Command{
	Use:     "howamidoing",
	Aliases: []string{"h"},
	Short:   "Show accuracy of historical estimates",
	Long: `Show accuracy of historical estimates

est howamidoing

"How am I doing" provides a feedback loop for you to become a better estimator
by showing a visualization of the accuracy of your historical estimates.

The visualization is a dynamically generated PNG image in a temporary file,
automatically opened in the operating system's default viewer.
`,
	Run: func(cmd *cobra.Command, args []string) {
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			// TODO include only ratios in last N days to prevent chart from becoming unreadable.
			core.AccuracyRatioChart(ef.HistoricalEstimateAccuracyRatios(), time.Now())
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

func init() {
	rootCmd.AddCommand(howamidoingCmd)
}
