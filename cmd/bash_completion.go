package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const bashCompletionHelpMsg string = `To enable bash command completion for est, add this to your .bashrc or .bash_profile:

"""
# The next line enables bash command completion for est.
if [ ! -z $(which est) ]; then eval "$(est bash --code)"; fi
"""
`

// bashCompletionCmd represents the bash-completion command
var bashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Enable bash completion for est",
	Long:  bashCompletionHelpMsg,
	Run: func(cmd *cobra.Command, args []string) {
		if outputBashCompletionCode {
			rootCmd.GenBashCompletion(os.Stdout)
		} else {
			os.Stdout.WriteString(bashCompletionHelpMsg)
		}
	},
}

var outputBashCompletionCode bool

func init() {
	bashCompletionCmd.PersistentFlags().BoolVarP(&outputBashCompletionCode, "code", "c", false, "output bash completion code")
	rootCmd.AddCommand(bashCompletionCmd)
}
