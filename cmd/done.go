package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

// TODO add --ago <time> to allow task to be marked as done as of <time> ago. The idea here is est's automatic time-tracking shouldn't penalize you for being AFK and unable to mark a task as completed. <time> should probably be wall-clock time, if user starts at 3hr task at 9am on a Friday, finishes it at 1pm, and then doesn't mark it done until Monday at 1pm, then user would think of `--ago 3d`, and not think of business hours/days. Wall clock time is also more compatible with future version of est which may support non-business-day mode.
var doneCmd = &cobra.Command{
	Use:     "done",
	Aliases: []string{"d"},
	Short:   "Mark a task as done",
	Long: `Done - mark a task as done

est done <task ID prefix>

Mark a started task as done. To specify the task to start, use a prefix of the
task ID shown in 'est ls'.

The actual time spent on the task is calculated automatically by est.
TODO concurrent tasks share passage of time
TODO explain time calculation
TODO --ago <time>
TODO currently tasks can be restarted after marked done; finalize / explain

The (estimated hours, actual hours) for done tasks are used as data points to
predict delivery schedule of future tasks in 'est schedule'.

Examples:
  # Mark the task with ID prefix "3c" as done.
  est d 3c

  # Mark the task with ID prefix "8d6d9" as done.
  est d 8d6d9
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
			if !ef.Tasks[i].IsStarted() {
				fmt.Println("fatal: this task must be started before marking it done")
				os.Exit(1)
				return
			}
			ef.Tasks[i].Stop(time.Now())
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
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
	rootCmd.AddCommand(doneCmd)
}
