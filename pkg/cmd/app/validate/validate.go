package validate

import (
	"github.com/frontierdigital/sheriff/pkg/util/config"
	"github.com/frontierdigital/utils/output"
)

func Validate(configDir string) error {
	output.PrintlnfInfo("Loading and validating config from %s", configDir)
	config, err := config.LoadConfig(configDir)
	if err != nil {
		return err
	}

	return config.Validate()
}
