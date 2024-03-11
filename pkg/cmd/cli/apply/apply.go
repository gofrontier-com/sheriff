package apply

import (
	"github.com/spf13/cobra"

	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/groups"
	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/resources"
)

// NewCmdApply creates a command to apply config
func NewCmdApply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply config",
	}

	cmd.AddCommand(groups.NewCmdApplyGroups())
	cmd.AddCommand(resources.NewCmdApplyResources())

	return cmd
}
