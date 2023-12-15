package azure_rm_config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/frontierdigital/sheriff/pkg/core"
	"gopkg.in/yaml.v2"
)

func loadRoleManagementPolicyRulesets(policiesDirPath string) ([]*core.RoleManagementPolicyRuleset, error) {
	var roleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset

	entries, err := os.ReadDir(policiesDirPath)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		filePath := filepath.Join(policiesDirPath, e.Name())
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		roleManagementPolicyProperties := armauthorization.RoleManagementPolicyProperties{}
		err = roleManagementPolicyProperties.UnmarshalJSON(yamlFile)
		if err != nil {
			return nil, err
		}

		roleManagementPolicyRuleset := &core.RoleManagementPolicyRuleset{
			Name: strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
		}
		roleManagementPolicyRuleset.Rules = append(roleManagementPolicyRuleset.Rules, roleManagementPolicyProperties.Rules...)

		roleManagementPolicyRulesets = append(roleManagementPolicyRulesets, roleManagementPolicyRuleset)
	}

	return roleManagementPolicyRulesets, err
}

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

	roleManagementPolicyRulesets, err := loadRoleManagementPolicyRulesets(filepath.Join(configDirPath, "policies"))
	if err != nil {
		return nil, err
	}

	configurationData := core.AzureRmConfig{
		Groups:                       groups,
		RoleManagementPolicyRulesets: roleManagementPolicyRulesets,
		Users:                        users,
	}

	return &configurationData, nil
}
