package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const promptHelpMsg string = `est-prompt

est ships with a separate executable, est-prompt, which shows a summary of
started tasks in your prompt.

est-prompt uses a fixed-width format designed to be minimally distracting. It's
hardcoded to display in bold yellow which you can change in promptColor().

Add est-prompt to your bash prompt by adding this snippet into your PS1 variable:

"""
$(est-prompt 2> /dev/null)
"""
`

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Integrate est into your bash prompt",
	Long:  promptHelpMsg,
	Run: func(cmd *cobra.Command, args []string) {
		os.Stdout.WriteString(promptHelpMsg)
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
