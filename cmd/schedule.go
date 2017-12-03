package cmd

import (
	"os"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

// scheduleCmd represents the schedule command
var scheduleCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"s"},
	Short:   "Display a probabilistic schedule for unstarted, estimated tasks",
	Long: `Display a probabilistic schedule for unstarted, estimated tasks.

The schedule predicts the date on which all tasks will be delivered.

The prediction is based on a monte carlo simulation of how long future tasks will
actually take, based on personalized accuracy of historical task estimates.`,
	Run: func(cmd *cobra.Command, args []string) {
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			ts := ef.Tasks.NotDeleted().Estimated().NotStarted()
			dates := core.DeliverySchedule(ef.HistoricalEstimateAccuracyRatios(), ts)
			os.Stdout.WriteString(core.RenderDeliverySchedule(dates))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
		})
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}
