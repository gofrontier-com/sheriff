package plan

import (
	"github.com/spf13/cobra"

	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/groups"
	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/resources"
)

// NewCmdApply creates a command to plan changes
func NewCmdPlan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan changes",
	}

	cmd.AddCommand(groups.NewCmdPlanGroups())
	cmd.AddCommand(resources.NewCmdPlanResources())

	return cmd
}
