package validate

import (
	"github.com/frontierdigital/sheriff/pkg/util/azure_rbac_config"
	"github.com/frontierdigital/utils/output"
)

func ValidateAzureRbac(configDir string) error {
	output.PrintlnfInfo("Loading and validating Azure Rbac config from %s", configDir)
	config, err := azure_rbac_config.Load(configDir)
	if err != nil {
		return err
	}

	return config.Validate()
}
