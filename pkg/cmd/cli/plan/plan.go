package plan

import (
	"github.com/spf13/cobra"
)

// NewCmdApply creates a command to plan changes
func NewCmdPlan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan changes",
	}

	cmd.AddCommand(NewCmdPlanAzureRm())

	return cmd
}
