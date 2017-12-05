package cmd

import (
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"s"},
	Short:   "Display a probabilistic schedule for unstarted, estimated tasks",
	Long: `Display a probabilistic schedule for unstarted, estimated tasks.

The schedule predicts the date on which all tasks will be delivered.

The prediction is based on a monte carlo simulation of how long future tasks will
actually take, based on personalized accuracy of historical task estimates.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSchedule()
	},
}

var scheduleDisplayChart bool

func runSchedule() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.NotDeleted().Estimated().NotStarted().NotDone()
		dates := core.DeliverySchedule(ef.HistoricalEstimateAccuracyRatios(), ts)
		if scheduleDisplayChart {
			now := time.Now()
			dd := make([]float64, len(dates))
			for i := range dates {
				dd[i] = dates[i].Sub(now).Hours() / 24
			}
			core.PredictedDeliveryDateChart(dd, ts)
		} else {
			os.Stdout.WriteString(core.RenderDeliverySchedule(dates))
		}
	}, func() {
		// failed to load estconfig or estfile. Err printed elsewhere.
		os.Exit(1)
	})
}

func init() {
	scheduleCmd.PersistentFlags().BoolVarP(&scheduleDisplayChart, "chart", "c", false, "display schedule in a chart, press Q to quit")
	rootCmd.AddCommand(scheduleCmd)
}
