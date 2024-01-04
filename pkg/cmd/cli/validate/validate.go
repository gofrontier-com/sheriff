package validate

import (
	"github.com/spf13/cobra"
)

// NewCmdValidate creates a command to validate config
func NewCmdValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config",
	}

	cmd.AddCommand(NewCmdValidateAzureRm())

	return cmd
}
