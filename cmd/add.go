package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a task",
	Long: `Add a new task

est add [(-e | --estimate) <estimate>] <task name>

The new task name is the concatenation of all non-flag args, no quotes required.
An estimate can be provided with -e, otherwise the new task will be unestimated.

// TODO move this estimate blurb to shared const between est-add and est-est
Estimates can be provided in hours "3.5h", days "2d", or weeks "0.75w". Defaults
to hours if unit unspecified. One day is equal to eight hours. One week is equal
to five days. In future, adherence to business days / hours may be configurable.

Examples:
  # Add an unestimated task named "my new task".
  est add my new task

  # Add an unestimated task named "my; new task is great".
  est add "my; new" task "is great"

  # Add an estimated task; the estimate unit defaults to hours.
  est add --estimate 2 change color to red

  # Add an estimated task; estimate is half a day; flag can go after task name.
  est add fix the bug -e 0.5d
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		fmt.Println("add called")
		fmt.Printf("args %d %+v\n", len(args), args)
		fmt.Printf("name %s\n", name)
	},
}

var addCmdNewTaskEstimate string

func init() {
	addCmd.PersistentFlags().StringVarP(&addCmdNewTaskEstimate, "estimate", "e", "", "estimate new task")
	rootCmd.AddCommand(addCmd)
}
