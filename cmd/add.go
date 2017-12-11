package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a task",
	Long: `Add a new task

est add [(-e | --estimate) <estimate>] [--start] <task name>

The new task name is the concatenation of all non-flag args, no quotes required.
An estimate can be provided with -e, otherwise the new task will be unestimated.
If an estimate was provided, the new task can be immediately started with -s.

Estimates can be provided in minutes "30m", hours "3.5h", days "2d", or weeks
"0.75w". Defaults to hours if unit unspecified. One day is equal to eight hours.
One week is equal to five days. In future, adherence to business days / hours
may be configurable.

The start time can be in the past with -a, using the same duration syntax as -e.

Examples:
  # Add an unestimated task named "my new task".
  est a my new task

  # Add an unestimated task named "my; new task is great".
  est a "my; new" task "is great"

  # Add an estimated task; the estimate unit defaults to hours.
  est a --estimate 2 change color to red

  # Add an estimated task; estimate is half a day; flag can go after task name.
  est a fix the bug -e 0.5d

  # Add and start an estimated task.
  est a -e 30m -s add another button

  # Add an estimated task and start it as of one hour ago.
  est a -e 4h -s -a 1h "this is a four hour task I started an hour ago"
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.TrimSpace(strings.Join(args, " "))
		if len(name) < 1 {
			fmt.Println("fatal: no task name given")
			os.Exit(1)
			return
		}
		estimate, err := parseDurationHours(addCmdEstimate, "estimate")
		if err != nil {
			fmt.Println("fatal: " + err.Error())
			os.Exit(1)
			return
		}
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			t := core.NewTask()
			if err := t.SetName(name); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			if estimate != 0 {
				if err := t.SetEstimated(estimate); err != nil {
					fmt.Printf("fatal: %v\n", err)
					os.Exit(1)
					return
				}
			}
			if addCmdStartNow && estimate == 0 {
				fmt.Println("fatal: cannot immediately start new task because no estimate was given")
				os.Exit(1)
				return
			}
			ef.Tasks = append(ef.Tasks, t)
			if addCmdStartNow {
				startTime := applyFlagAgo(time.Now())
				if err := ef.Tasks.Start(len(ef.Tasks)-1, startTime); err != nil {
					fmt.Printf("fatal: %v\n", err)
					os.Exit(1)
					return
				}
			}
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			fmt.Println(core.RenderTaskOneLineSummary(t))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

var addCmdEstimate string
var addCmdStartNow bool

func init() {
	addCmd.PersistentFlags().StringVarP(&addCmdEstimate, "estimate", "e", "", "estimate new task")
	addCmd.PersistentFlags().StringVarP(&flagAgo, "ago", "a", "", "when used with start, start duration ago from now")
	addCmd.PersistentFlags().BoolVarP(&addCmdStartNow, "start", "s", false, "immediately start new task")
	rootCmd.AddCommand(addCmd)
}
