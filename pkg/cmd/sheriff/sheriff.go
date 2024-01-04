package sheriff

import (
	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/apply"
	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/plan"
	"github.com/gofrontier-com/sheriff/pkg/cmd/cli/validate"
	vers "github.com/gofrontier-com/sheriff/pkg/cmd/cli/version"
	"github.com/spf13/cobra"
)

func NewRootCmd(version string, commit string, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                   "sheriff",
		DisableFlagsInUseLine: true,
		Short:                 "Sheriff is a command line tool to manage Azure role-based access control (Azure RBAC) and Microsoft Entra Priviliged Identity Management (Microsoft Entra PIM) using desired state configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.AddCommand(apply.NewCmdApply())
	rootCmd.AddCommand(plan.NewCmdPlan())
	rootCmd.AddCommand(validate.NewCmdValidate())
	rootCmd.AddCommand(vers.NewCmdVersion(version, commit, date))

	return rootCmd
}
