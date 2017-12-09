package cmd

import (
	"fmt"
	"os"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete a task",
	Long: `Delete a task

est rm <task ID prefix>

Delete an existing task. To specify the task to delete, use a prefix of the task
ID shown in 'est ls'.

Tasks are soft deleted by setting a flag on the task. You can restore deleted
tasks by editing your estfile directly and toggling this flag.

Deleted tasks are not used as prediction data in 'est schedule'.

Examples:
  # Delete the task with ID prefix "3c".
  est rm 3c

  # Delete the task with ID prefix "8d6d9".
  est e 8d6d9 1.5d
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("usage: est rm <task ID prefix>")
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
			if err := ef.Tasks[i].Delete(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
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
	rootCmd.AddCommand(rmCmd)
}
