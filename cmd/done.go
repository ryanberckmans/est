package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:     "done",
	Aliases: []string{"d"},
	Short:   "Mark a task as done",
	Long: `Done - mark a task as done

est done <task ID prefix>

Mark a started task as done. To specify the task to start, use a prefix of the
task ID shown in 'est ls'.

A task which is done can be restarted.

The actual time spent on the task is calculated automatically by est. For tasks
performed mostly during working hours, there is no need to log time on task,
est will do this automatically. For tasks performed mostly outside of working
hours, see 'est log'.

est's auto time tracking uses a customizable definition of working hours. For
now, working hours are hardcoded in cmd/root.go.

When multiple tasks are started, est will share the passage of time equally
among all started tasks. For example, if two tasks are started and time passes
from 9am to 10am during working hours, then each task will receive half the
elapsed duration, which is thirty minutes.

The time at which a task is marked done can be in the past with -a, using the
same duration syntax as 'est estimate'.

The (estimated hours, actual hours) for done tasks are used as data points to
predict delivery schedule of future tasks in 'est schedule'.

Examples:
  # Mark the task with ID prefix "3c" as done.
  est d 3c

  # Mark the task with ID prefix "8d6d9" as done.
  est d 8d6d9

  # Mark the task with ID prefix "57" as done as of two and half hours ago.
  est d -a 2.5h 57
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("usage: est done <task ID prefix>")
			os.Exit(1)
			return
		}
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			i := ef.Tasks.FindByIDPrefix(args[0])
			if i < 0 {
				fmt.Printf("fatal: no task with ID prefix '%s'\n", args[0])
				os.Exit(1)
				return
			}
			doneTime := applyFlagAgo(time.Now())
			doFlagLog(ef.Tasks[i], doneTime)
			if err := ef.Tasks.Done(globalWorkTimes, i, doneTime); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			fmt.Println(core.RenderTaskOneLineSummary(ef.Tasks[i], true))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

func init() {
	doneCmd.PersistentFlags().StringVarP(&flagLog, "log", "l", "", "log time worked prior to marking this task as done (overrides auto time tracking)")
	doneCmd.PersistentFlags().StringVarP(&flagAgo, "ago", "a", "", "done duration ago from now")
	rootCmd.AddCommand(doneCmd)
}
