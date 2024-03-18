package resources

import (
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/util/resources_config"
)

func ValidateResources(configDir string) error {
	output.PrintlnInfo("Initialising...")

	output.PrintlnInfo("- Loading and validating config\n")

	config, err := resources_config.Load(configDir)
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
