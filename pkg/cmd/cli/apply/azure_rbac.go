package apply

import (
	"os"

	"github.com/frontierdigital/sheriff/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

var (
	configDir      string
	dryRun         bool
	subscriptionId string
)

// NewCmdApplyAzureRbac creates a command to apply the Azure Rbac config
func NewCmdApplyAzureRbac() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerbac",
		Short: "Apply Azure Rbac config",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := apply.ApplyAzureRbac(configDir, subscriptionId, dryRun); err != nil {
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
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry-run")
	cmd.Flags().StringVarP(&subscriptionId, "subscription-id", "s", "", "Subscription Id") // TODO: Support name

	cobra.MarkFlagRequired(cmd.Flags(), "subscription-id")

	return cmd
}