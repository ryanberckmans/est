package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"c"},
	Short:   "Display a probabilistic schedule for unstarted, estimated tasks",
	Long: `Display a probabilistic schedule for unstarted, estimated tasks.

The schedule predicts the date on which all tasks will be delivered.

The prediction is based on a monte carlo simulation of how long future tasks will
actually take, based on personalized accuracy of historical task estimates.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSchedule()
	},
}

var scheduleDisplayDatesOnly bool

func runSchedule() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.IsNotDeleted().IsEstimated().IsNotStarted().IsNotDone()
		dates := core.DeliverySchedule(ef.HistoricalEstimateAccuracyRatios(), ts)
		ss := core.RenderDeliverySchedule(dates)
		if scheduleDisplayDatesOnly {
			s := strings.Join(ss[:], "\n") + "\n"
			os.Stdout.WriteString(s)
		} else {
			now := time.Now()
			// Convert dates into wall clock days in future, because termui supports only float data
			dd := make([]float64, len(dates))
			for i := range dates {
				dd[i] = dates[i].Sub(now).Hours() / 24
			}
			core.PredictedDeliveryDateChart(dd, ts, ss[:])
		}
	}, func() {
		// failed to load estconfig or estfile. Err printed elsewhere.
		os.Exit(1)
	})
}

func init() {
	scheduleCmd.PersistentFlags().BoolVarP(&scheduleDisplayDatesOnly, "dates-only", "d", false, "display dates only, no chart, non-interactively")
	rootCmd.AddCommand(scheduleCmd)
}
