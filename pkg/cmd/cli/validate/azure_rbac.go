package validate

import (
	"os"

	"github.com/frontierdigital/sheriff/pkg/cmd/app/validate"
	"github.com/spf13/cobra"
)

var (
	configDir string
)

// NewCmdValidate creates a command to validate the Azure Rm config
func NewCmdValidateAzureRm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerm",
		Short: "Validate Azure RM config",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := validate.ValidateAzureRm(configDir); err != nil {
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
