package resources

import (
	"os"

	"github.com/gofrontier-com/sheriff/pkg/cmd/app/resources"
	"github.com/spf13/cobra"
)

// NewCmdValidateResources creates a command to validate the Azure Resources config
func NewCmdValidateResources() *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Validate Azure Resources config",
		RunE: func(_ *cobra.Command, _ []string) error {
			printHeader("Validate", configDir, nil)

			if err := resources.ValidateResources(configDir); err != nil {
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
