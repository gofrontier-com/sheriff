package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/frontierdigital/sheriff/pkg/core"
	"gopkg.in/yaml.v2"
)

func LoadConfig(configDirPath string) (*core.Config, error) {
	configurationData := core.Config{}

	entries, err := os.ReadDir(filepath.Join(configDirPath, "groups"))
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		filePath := filepath.Join(configDirPath, "groups", e.Name())
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var group core.Group

		err = yaml.Unmarshal(yamlFile, &group)
		if err != nil {
			return nil, err
		}

		group.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		configurationData.Groups = append(configurationData.Groups, &group)
	}
	return &configurationData, nil
}
