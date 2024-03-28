package groups_config

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
		if filepath.Ext(e.Name()) != ".yml" && filepath.Ext(e.Name()) != ".yaml" {
			continue
		}

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

		if roleManagementPolicyRuleset.Rules == nil {
			continue
		}

		roleManagementPolicyRuleset.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		for _, r := range roleManagementPolicyRuleset.Rules {
			r.Patch = convertPatchStruct(r.Patch)
		}

		roleManagementPolicyRulesets = append(roleManagementPolicyRulesets, &roleManagementPolicyRuleset)
	}

	return roleManagementPolicyRulesets, err
}

func loadPolicies(policiesDirPath string) ([]*core.GroupPolicy, error) {
	var policies []*core.GroupPolicy

	if _, err := os.Stat(policiesDirPath); err != nil {
		if os.IsNotExist(err) {
			return policies, nil
		}
	}

	entries, err := os.ReadDir(policiesDirPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if filepath.Ext(e.Name()) != ".yml" && filepath.Ext(e.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(policiesDirPath, e.Name())
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var policy core.GroupPolicy

		err = yaml.Unmarshal(yamlFile, &policy)
		if err != nil {
			return nil, err
		}

		if policy.Default == nil &&
			policy.ManagedGroups == nil {
			continue
		}

		policy.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		policies = append(policies, &policy)
	}

	return policies, nil
}

func loadPrincipals(principalsDirPath string) ([]*core.GroupPrincipal, error) {
	var principals []*core.GroupPrincipal

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
		if filepath.Ext(e.Name()) != ".yml" && filepath.Ext(e.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(principalsDirPath, e.Name())
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var principal core.GroupPrincipal

		err = yaml.Unmarshal(yamlFile, &principal)
		if err != nil {
			return nil, err
		}

		if principal.ManagedGroups == nil {
			continue
		}

		principal.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		principals = append(principals, &principal)
	}

	return principals, err
}

func Load(configDirPath string) (*core.GroupsConfig, error) {
	// errors := validateDirStructure(configDirPath)
	// if len(errors) > 0 {
	// 	return nil, fmt.Errorf("invalid config dir structure: %v", errors)
	// }

	groups, err := loadPrincipals(filepath.Join(configDirPath, "groups"))
	if err != nil {
		return nil, err
	}

	users, err := loadPrincipals(filepath.Join(configDirPath, "users"))
	if err != nil {
		return nil, err
	}

	roleManagementPolicyRulesets, err := loadRoleManagementPolicyRulesets(filepath.Join(configDirPath, "policies", "rulesets"))
	if err != nil {
		return nil, err
	}

	policies, err := loadPolicies(filepath.Join(configDirPath, "policies"))
	if err != nil {
		return nil, err
	}

	configurationData := core.GroupsConfig{
		Groups:   groups,
		Policies: policies,
		Rulesets: roleManagementPolicyRulesets,
		Users:    users,
	}

	if len(configurationData.Groups) == 0 && len(configurationData.Users) == 0 {
		return &configurationData, &core.ConfigurationEmptyError{}
	}

	return &configurationData, nil
}
