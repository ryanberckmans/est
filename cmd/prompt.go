package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const promptHelpMsg string = `est ships with a separate executable, est-prompt, which adds live est task information to your prompt.

est-prompt is designed to be opinionated and minimally distracting.

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
