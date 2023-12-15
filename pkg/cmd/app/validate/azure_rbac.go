package validate

import (
	"github.com/frontierdigital/sheriff/pkg/util/azure_rm_config"
	"github.com/frontierdigital/utils/output"
)

func ValidateAzureRm(configDir string) error {
	output.PrintlnInfo("Initialising...")

	output.PrintlnInfo("- Loading and validating config\n")

	config, err := azure_rm_config.Load(configDir)
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
