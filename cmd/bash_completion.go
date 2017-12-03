package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// bashCompletionCmd represents the bash-completion command
var bashCompletionCmd = &cobra.Command{
	Use:   "bash-completion",
	Short: "Instructions to enable bash completion for est",
	Long:  `To enable bash command completion for est, run "est bash-completion" and follow the displayed instructions.`,
	Run: func(cmd *cobra.Command, args []string) {
		if outputBashCompletionCode {
			rootCmd.GenBashCompletion(os.Stdout)
		} else {
			os.Stdout.WriteString(`To enable bash command completion for est, add this to your .bashrc or .bash_profile:
"""
# The next line enables bash command completion for est.
if [ ! -z $(which est) ]; then eval "$(est bash-completion --code)"; fi
"""
`)
		}
	},
}

var outputBashCompletionCode bool

func init() {
	bashCompletionCmd.PersistentFlags().BoolVarP(&outputBashCompletionCode, "code", "c", false, "output bash completion code")
	rootCmd.AddCommand(bashCompletionCmd)
}
