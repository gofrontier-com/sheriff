package validate

import (
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/util/azurerm_config"
)

func ValidateAzureRm(configDir string) error {
	output.PrintlnInfo("Initialising...")

	output.PrintlnInfo("- Loading and validating config\n")

	config, err := azurerm_config.Load(configDir)
	if err != nil {
		return err
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	output.PrintlnInfo("Configuration is valid!\n")

	return nil
}
