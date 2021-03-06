package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Aliases: []string{"l"},
	Short:   "Log time worked on a task (in lieu of auto time tracking)",
	Long: `Log time worked on a task (in lieu of auto time tracking)

est log <task ID prefix> <duration>

Most users should not use 'est log' and instead rely on auto time tracking. See
'est help done' for an explanation of auto time tracking and how it interacts
with 'est log'.

'est log' is provided as an escape hatch. In particular, auto time tracking
poorly handles work done entirely outside of business hours.

As a rule of thumb, use 'est log' if more than half of actual task hours are
done outside of working hours. Otherwise rely on auto time tracking.

To specify the task on which to log time, use a prefix of the task ID shown in
'est ls'.

'est log' can be used any number of times on a started task; successive
logged durations will be added to the task's actual hours.

'est add', 'est start', and 'est done' take --log which can be used to
avoid running 'est log' directly.

The logged duration uses the same syntax as 'est estimate', and can be provided
as second argument or as -l.

Examples:
  # Log 7 hours worked on the task with ID prefix "3c".
  est l 3c 7h

  # Log 45 minutes worked on the task with ID prefix "8d6d9".
  est l 8d6d9 45m

  # Log 45 minutes worked on the task with ID prefix "94".
  est -l 0.75h 94
`,
	Run: func(cmd *cobra.Command, args []string) {
		if flagLog == "" && len(args) > 1 {
			// est log can take either <log> or --log <duration>
			flagLog = args[1]
		}
		if flagLog == "" {
			fmt.Println("usage: est log <task ID prefix> [-l] <duration>")
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
			doFlagLog(ef.Tasks[i], time.Now())
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
	logCmd.PersistentFlags().StringVarP(&flagLog, "log", "l", "", "log time worked")
	rootCmd.AddCommand(logCmd)
}
