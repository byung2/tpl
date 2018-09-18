package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func newCompletionCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "completion",
		Short: "Emit bash completion",
		Long: `Emit bash completion
										
Installation (Linux):
  sudo sh -c "tpl completion > /etc/bash_completion.d/tpl" \
	&& source /etc/bash_completion.d/tpl
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Root().GenBashCompletion(os.Stdout)
		},
	}
	return createCmd
}
