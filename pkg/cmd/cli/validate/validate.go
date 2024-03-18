package validate

import (
	"github.com/spf13/cobra"

	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/resources"
)

// NewCmdValidate creates a command to validate config
func NewCmdValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config",
	}

	cmd.AddCommand(resources.NewCmdValidateResources())

	return cmd
}
