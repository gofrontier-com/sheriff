package groups

import (
	"os"

	"github.com/gofrontier-com/sheriff/pkg/cmd/app/groups"
	"github.com/spf13/cobra"
)

// NewCmdPlanGroups creates a command to plan the Groups config changes
func NewCmdPlanGroups() *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Plan Groups config changes",
		RunE: func(c *cobra.Command, _ []string) error {
			printHeader("Plan", configDir)

			if err := groups.ApplyGroups(configDir, true); err != nil {
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

	return cmd
}
