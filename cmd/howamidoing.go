package cmd

import (
	"fmt"
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

Shows last 90 days of history.
`,
	Run: func(cmd *cobra.Command, args []string) {
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			now := time.Now()
			ars := ef.HistoricalEstimateAccuracyRatios().After(now.Add(-time.Hour * 24 * 90)) // show only 90 days of history to ensure chart is readable and history eventually drops off (hopefully estimator improves)
			if err := core.AccuracyRatioChart(ars, now); err != nil {
				fmt.Println("fatal: " + err.Error())
				os.Exit(1)
				return
			}
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

func init() {
	rootCmd.AddCommand(howamidoingCmd)
}
