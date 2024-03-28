package groups

import (
	"os"

	"github.com/gofrontier-com/sheriff/pkg/cmd/app/groups"
	"github.com/spf13/cobra"
)

// NewCmdApplyGroups creates a command to apply the Groups config
func NewCmdApplyGroups() *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Apply Groups config",
		RunE: func(_ *cobra.Command, _ []string) error {
			var action string
			if planOnly {
				action = "Apply (plan-only)"
			} else {
				action = "Apply"
			}

			printHeader(action, configDir)

			if err := groups.ApplyGroups(configDir, planOnly); err != nil {
				return err
			}

			return nil
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&configDir, "config-dir", "c", wd, "Config directory")
	cmd.Flags().BoolVarP(&planOnly, "plan-only", "p", false, "Plan-only")

	return cmd
}
