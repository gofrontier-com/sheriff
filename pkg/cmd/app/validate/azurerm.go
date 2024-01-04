package validate

import (
	"github.com/frontierdigital/sheriff/pkg/util/azurerm_config"
	"github.com/frontierdigital/utils/output"
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
