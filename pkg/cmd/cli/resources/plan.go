package resources

import (
	"os"

	"github.com/gofrontier-com/sheriff/pkg/cmd/app/resources"
	"github.com/spf13/cobra"
)

// NewCmdPlanResources creates a command to plan the Azure Resources config changes
func NewCmdPlanResources() *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Plan Azure Resources config changes",
		RunE: func(c *cobra.Command, _ []string) error {
			printHeader("Plan", configDir, &subscriptionId)

			if err := resources.ApplyResources(configDir, subscriptionId, true); err != nil {
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
	cmd.Flags().StringVarP(&subscriptionId, "subscription-id", "s", "", "Subscription Id") // TODO: Support name

	cobra.MarkFlagRequired(cmd.Flags(), "subscription-id")

	return cmd
}
