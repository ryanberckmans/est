package cmd

import (
	"fmt"
	"os"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var estimateCmd = &cobra.Command{
	Use:     "estimate",
	Aliases: []string{"e"},
	Short:   "Estimate a task",
	Long: `Estimate a task

est estimate <task ID prefix> <estimate>

Change the estimate on an existing task. To specify the task to estimate, use a
prefix of the task ID shown in 'est ls'.

Estimates can be provided in hours "3.5h", days "2d", or weeks "0.75w". Defaults
to hours if unit unspecified. One day is equal to eight hours. One week is equal
to five days. In future, adherence to business days / hours may be configurable.

Examples:
  # Estimate the task with ID prefix "3c" at 7 hours.
  est e 3c 7

  # Estimate the task with ID prefix "8d6d9" at 1.5 days.
  est e 8d6d9 1.5d

  # Estimate the task with ID prefix "94" at 0.25 weeks.
  est e 94 0.25w
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("usage: est estimate <task ID prefix> <estimate>")
			os.Exit(1)
			return
		}
		estimate, err := parseEstimate(args[1])
		if err != nil {
			fmt.Println("fatal: " + err.Error())
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
			if len(ef.Tasks[i].Hours) < 1 {
				ef.Tasks[i].Hours = []float64{estimate}
			} else {
				ef.Tasks[i].Hours[0] = estimate
			}
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
	rootCmd.AddCommand(estimateCmd)
}