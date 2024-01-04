package azurerm_config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/frontierdigital/sheriff/pkg/core"
	"gopkg.in/yaml.v2"
)

func loadRoleManagementPolicyPatches(patchesDirPath string) ([]*core.RoleManagementPolicyPatch, error) {
	var roleManagementPolicyPatches []*core.RoleManagementPolicyPatch

	if _, err := os.Stat(patchesDirPath); err != nil {
		if os.IsNotExist(err) {
			return roleManagementPolicyPatches, nil
		}
	}

	entries, err := os.ReadDir(patchesDirPath)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		filePath := filepath.Join(patchesDirPath, e.Name())
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var roleManagementPolicyPatch core.RoleManagementPolicyPatch

		err = yaml.Unmarshal(yamlFile, &roleManagementPolicyPatch)
		if err != nil {
			return nil, err
		}

		roleManagementPolicyPatch.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		roleManagementPolicyPatches = append(roleManagementPolicyPatches, &roleManagementPolicyPatch)
	}

	return roleManagementPolicyPatches, err
}

func loadPrincipals(principalsDirPath string) ([]*core.Principal, error) {
	var principals []*core.Principal

	if _, err := os.Stat(principalsDirPath); err != nil {
		if os.IsNotExist(err) {
			return principals, nil
		}
	}

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

	roleManagementPolicyPatches, err := loadRoleManagementPolicyPatches(filepath.Join(configDirPath, "policies"))
	if err != nil {
		return nil, err
	}

	configurationData := core.AzureRmConfig{
		Groups:                      groups,
		RoleManagementPolicyPatches: roleManagementPolicyPatches,
		Users:                       users,
	}

	if len(configurationData.Groups) == 0 && len(configurationData.Users) == 0 {
		return &configurationData, &core.ConfigurationEmptyError{}
	}

	return &configurationData, nil
}
