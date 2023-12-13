package apply

import (
	"os"

	"github.com/frontierdigital/sheriff/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

var (
	configDir      string
	planOnly       bool
	subscriptionId string
)

// NewCmdApplyAzureRm creates a command to apply the Azure RM config
func NewCmdApplyAzureRm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerm",
		Short: "Apply Azure Resource Manager config",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := apply.ApplyAzureRm(configDir, subscriptionId, planOnly); err != nil {
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
