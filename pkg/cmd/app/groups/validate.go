package groups

import (
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/util/groups_config"
)

func ValidateGroups(configDir string) error {
	output.PrintlnInfo("Initialising...")

	output.PrintlnInfo("- Loading and validating config\n")

	config, err := groups_config.Load(configDir)
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
