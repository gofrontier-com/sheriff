package apply

import (
	"github.com/spf13/cobra"
)

// NewCmdApply creates a command to apply config
func NewCmdApply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply config",
	}

	cmd.AddCommand(NewCmdApplyAzureRm())

	return cmd
}
