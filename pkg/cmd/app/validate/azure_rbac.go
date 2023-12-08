package validate

import (
	"github.com/frontierdigital/sheriff/pkg/util/azure_rm_config"
	"github.com/frontierdigital/utils/output"
)

func ValidateAzureRm(configDir string) error {
	output.PrintlnfInfo("Loading and validating Azure RM config from %s", configDir)
	config, err := azure_rm_config.Load(configDir)
	if err != nil {
		return err
	}

	return config.Validate()
}
