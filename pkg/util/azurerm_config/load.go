package azurerm_config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrontier-com/sheriff/pkg/core"
	"gopkg.in/yaml.v2"
)

func convertPatchStruct(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertPatchStruct(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertPatchStruct(v)
		}
	}
	return i
}

func loadRoleManagementPolicyRulesets(patchesDirPath string) ([]*core.RoleManagementPolicyRuleset, error) {
	var roleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset

	if _, err := os.Stat(patchesDirPath); err != nil {
		if os.IsNotExist(err) {
			return roleManagementPolicyRulesets, nil
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

		var roleManagementPolicyRuleset core.RoleManagementPolicyRuleset

		err = yaml.Unmarshal(yamlFile, &roleManagementPolicyRuleset)
		if err != nil {
			return nil, err
		}

		roleManagementPolicyRuleset.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		for _, r := range roleManagementPolicyRuleset.Rules {
			r.Patch = convertPatchStruct(r.Patch)
		}

		roleManagementPolicyRulesets = append(roleManagementPolicyRulesets, &roleManagementPolicyRuleset)
	}

	return roleManagementPolicyRulesets, err
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

	roleManagementPolicyRulesets, err := loadRoleManagementPolicyRulesets(filepath.Join(configDirPath, "rulesets"))
	if err != nil {
		return nil, err
	}

	configurationData := core.AzureRmConfig{
		Groups:   groups,
		Rulesets: roleManagementPolicyRulesets,
		Users:    users,
	}

	if len(configurationData.Groups) == 0 && len(configurationData.Users) == 0 {
		return &configurationData, &core.ConfigurationEmptyError{}
	}

	return &configurationData, nil
}
