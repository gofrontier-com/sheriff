package azure_rm_config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/frontierdigital/sheriff/pkg/core"
	"gopkg.in/yaml.v2"
)

func loadPrincipals(principalsDirPath string) ([]*core.Principal, error) {
	var principals []*core.Principal

	entries, err := os.ReadDir(principalsDirPath)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		filePath := filepath.Join(principalsDirPath, e.Name())
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var principal core.Principal

		err = yaml.Unmarshal(yamlFile, &principal)
		if err != nil {
			return nil, err
		}

		principal.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		principals = append(principals, &principal)
	}

	return principals, err
}

func Load(configDirPath string) (*core.AzureRmConfig, error) {
	groups, err := loadPrincipals(filepath.Join(configDirPath, "groups"))
	if err != nil {
		return nil, err
	}

	users, err := loadPrincipals(filepath.Join(configDirPath, "users"))
	if err != nil {
		return nil, err
	}

	configurationData := core.AzureRmConfig{
		Groups: groups,
		Users:  users,
	}

	return &configurationData, nil
}
