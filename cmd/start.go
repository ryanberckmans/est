package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

// TODO add --time to allow task to be started as of a different time than now
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a task",
	Long: `Start a task

est start <task ID prefix>

Start an existing task. To specify the task to start, use a prefix of the task
ID shown in 'est ls'.

A summary of started tasks is shown in the est prompt. See 'est help prompt'.

Examples:
  # Start the task with ID prefix "3c".
  est s 3c

  # Start the task with ID prefix "8d6d9".
  est s 8d6d9
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
			if !ef.Tasks[i].IsEstimated() {
				fmt.Println("fatal: this task must be estimated before starting it")
				os.Exit(1)
				return
			}
			if ef.Tasks[i].IsStarted() {
				fmt.Println("fatal: this task is already started")
				os.Exit(1)
				return
			}
			ef.Tasks[i].Start(time.Now())
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
	rootCmd.AddCommand(startCmd)
}
