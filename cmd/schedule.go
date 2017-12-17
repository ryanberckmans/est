package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Display a predicted, probabilistic schedule for unstarted, estimated tasks",
	Long: `Display a predicted, probabilistic schedule for unstarted, estimated tasks.

The schedule predicts the date on which all tasks will be delivered.

The prediction is based on a monte carlo simulation of how long future tasks
will actually take, based on personalized accuracy of historical task estimates.

Personalized historical task estimates are partially faked if less than twenty
tasks are done.

The prediction is most accurate when most of your tasks' estimated and actual
hours are under 16 hours; your history of done tasks numbers in the dozens;
and you track the majority of your working time in est.

The prediction is most useful if your expected future tasks are estimated in
est, because those are the tasks predicted by 'est schedule'. If many of your
future tasks are not estimated in est, or the tasks estimated in est are never
actually worked on, then the usefulness of 'est schedule' will be reduced.

In future, 'est schedule' should allow selection of tasks to estimate, e.g.
all tasks related to one project, and merging of estfiles for team scheduling.
`,
	Run: func(cmd *cobra.Command, args []string) {
		runSchedule()
	},
}

var scheduleDisplayDatesOnly bool

func runSchedule() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.IsNotDeleted().IsEstimated().IsNotStarted().IsNotDone()
		now := time.Now()
		os.Stdout.WriteString("Predicting delivery schedule for unstarted, estimated tasks...")
		dates := core.DeliverySchedule(globalWorkTimes, now, ef.HistoricalEstimateAccuracyRatios(), ts)
		ss := core.RenderDeliverySchedule(dates)
		os.Stdout.WriteString("done\n")
		if scheduleDisplayDatesOnly {
			s := strings.Join(ss[:], "\n") + "\n"
			os.Stdout.WriteString(s)
		} else {
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
