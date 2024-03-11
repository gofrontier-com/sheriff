package resources_config

import (
	"fmt"
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

func loadPolicies(policiesDirPath string) ([]*core.Policy, error) {
	var policies []*core.Policy

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

		var policy core.Policy

		err = yaml.Unmarshal(yamlFile, &policy)
		if err != nil {
			return nil, err
		}

		if policy.Default == nil &&
			policy.Subscription == nil &&
			policy.ResourceGroups == nil &&
			policy.Resources == nil {
			continue
		}

		policy.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		policies = append(policies, &policy)
	}

	return policies, nil
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
		if filepath.Ext(e.Name()) != ".yml" && filepath.Ext(e.Name()) != ".yaml" {
			continue
		}

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

		if principal.Subscription == nil &&
			principal.ResourceGroups == nil &&
			principal.Resources == nil {
			continue
		}

		principal.Name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

		principals = append(principals, &principal)
	}

	return principals, err
}

func validateDirStructure(configDirPath string) []error {
	errors := []error{}

	if _, err := os.Stat(configDirPath); err != nil {
		if os.IsNotExist(err) {
			return append(errors, fmt.Errorf("config dir path does not exist: %s", configDirPath))
		}
	}

	entries, err := os.ReadDir(configDirPath)
	if err != nil {
		return append(errors, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			errors = append(errors, fmt.Errorf("unexpected file in config dir: %s", e.Name()))
			continue
		}

		if e.Name() == "groups" || e.Name() == "users" {
			entries, err := os.ReadDir(filepath.Join(configDirPath, e.Name()))
			if err != nil {
				return append(errors, err)
			}
			for _, f := range entries {
				if f.IsDir() {
					errors = append(errors, fmt.Errorf("unexpected dir in %s: %s", e.Name(), f.Name()))
				}
			}

			continue
		}

		if e.Name() == "policies" {
			entries, err := os.ReadDir(filepath.Join(configDirPath, e.Name()))
			if err != nil {
				return append(errors, err)
			}
			for _, f := range entries {
				if f.IsDir() {
					if f.Name() == "rulesets" {
						entries, err := os.ReadDir(filepath.Join(configDirPath, e.Name(), f.Name()))
						if err != nil {
							return append(errors, err)
						}
						for _, g := range entries {
							if g.IsDir() {
								errors = append(errors, fmt.Errorf("unexpected dir in %s: %s", f.Name(), g.Name()))
							}
						}

						continue
					}

					errors = append(errors, fmt.Errorf("unexpected dir in %s: %s", e.Name(), f.Name()))
				}
			}

			continue
		}

		errors = append(errors, fmt.Errorf("unexpected dir in config dir: %s", e.Name()))
	}

	return errors
}

func Load(configDirPath string) (*core.ResourcesConfig, error) {
	errors := validateDirStructure(configDirPath)
	if len(errors) > 0 {
		return nil, fmt.Errorf("invalid config dir structure: %v", errors)
	}

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

	configurationData := core.ResourcesConfig{
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
