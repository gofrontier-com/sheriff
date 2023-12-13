package plan

import (
	"os"

	"github.com/frontierdigital/sheriff/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

var (
	configDir      string
	subscriptionId string
)

// NewCmdPlanAzureRm creates a command to llan the Azure RM config changes
func NewCmdPlanAzureRm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerm",
		Short: "Plan Azure Resource Manager config changes",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := apply.ApplyAzureRm(configDir, subscriptionId, true); err != nil {
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
