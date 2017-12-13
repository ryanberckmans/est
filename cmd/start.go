package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "Start a task",
	Long: `Start a task

est start <task ID prefix>

Start an existing task. To specify the task to start, use a prefix of the task
ID shown in 'est ls'.

A summary of started tasks is shown in the est prompt. See 'est help prompt'.

The start time can be in the past with -a, using the same duration syntax as the
estimate command.

Examples:
  # Start the task with ID prefix "3c".
  est s 3c

  # Start the task with ID prefix "8d6d9".
  est s 8d6d9

  # Start as of forty five minutes ago the task with ID Prefix "813".
  est s -a 45m 813
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("usage: est start <task ID prefix>")
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
			startTime := applyFlagAgo(time.Now())
			if err := ef.Tasks.Start(i, startTime); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			fmt.Println(core.RenderTaskOneLineSummary(ef.Tasks[i]))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

func init() {
	startCmd.PersistentFlags().StringVarP(&flagAgo, "ago", "a", "", "start duration ago from now")
	rootCmd.AddCommand(startCmd)
}
