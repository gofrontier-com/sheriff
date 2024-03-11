package resources

import (
	"os"

	"github.com/gofrontier-com/sheriff/pkg/cmd/app/resources"
	"github.com/spf13/cobra"
)

// NewCmdApplyResources creates a command to apply the Azure Resources config
func NewCmdApplyResources() *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Apply Azure Resources config",
		RunE: func(_ *cobra.Command, _ []string) error {
			var action string
			if planOnly {
				action = "Apply (plan-only)"
			} else {
				action = "Apply"
			}

			printHeader(action, configDir, &subscriptionId)

			if err := resources.ApplyResources(configDir, subscriptionId, planOnly); err != nil {
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
	cmd.Flags().StringVarP(&subscriptionId, "subscription-id", "s", "", "Subscription Id") // TODO: Support name

	cobra.MarkFlagRequired(cmd.Flags(), "subscription-id")

	return cmd
}
