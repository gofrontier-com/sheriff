package core

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/go-playground/validator/v10"
)

func (c *ResourcesConfig) GetGroupAssignmentSchedules(subscriptionId string) []*Schedule {
	return getResourcesAssignmentSchedules(c.Groups, subscriptionId)
}

func (c *ResourcesConfig) GetGroupEligibilitySchedules(subscriptionId string) []*Schedule {
	return getResourcesEligibilitySchedules(c.Groups, subscriptionId)
}

func (c *ResourcesConfig) GetPolicyByRoleName(roleName string) *ResourcePolicy {
	policy := linq.From(c.Policies).SingleWithT(func(p *ResourcePolicy) bool {
		return p.Name == roleName
	})

	if policy == nil {
		policy = linq.From(c.Policies).SingleWithT(func(p *ResourcePolicy) bool {
			return p.Name == "default"
		})
	}

	if p, ok := policy.(*ResourcePolicy); ok {
		return p
	} else {
		return nil
	}
}

func (c *ResourcesConfig) GetScopeRoleNameCombinations(subscriptionId string) []*TargetRoleNameCombination {
	groupAssignmentSchedules := c.GetGroupAssignmentSchedules(subscriptionId)
	userAssignmentSchedules := c.GetUserAssignmentSchedules(subscriptionId)
	groupEligibilitySchedules := c.GetGroupEligibilitySchedules(subscriptionId)
	userEligibilitySchedules := c.GetUserEligibilitySchedules(subscriptionId)

	allSchedules := append(groupAssignmentSchedules, userAssignmentSchedules...)
	allSchedules = append(allSchedules, groupEligibilitySchedules...)
	allSchedules = append(allSchedules, userEligibilitySchedules...)

	var scopeRoleNameCombinations []*TargetRoleNameCombination
	linq.From(allSchedules).SelectT(func(s *Schedule) *TargetRoleNameCombination {
		return &TargetRoleNameCombination{
			RoleName: s.RoleName,
			Target:   s.Target,
		}
	}).DistinctByT(func(s *TargetRoleNameCombination) string {
		return fmt.Sprintf("%s:%s", s.Target, s.RoleName)
	}).ToSlice(&scopeRoleNameCombinations)

	return scopeRoleNameCombinations
}

func (c *ResourcesConfig) GetUserAssignmentSchedules(subscriptionId string) []*Schedule {
	return getResourcesAssignmentSchedules(c.Users, subscriptionId)
}

func (c *ResourcesConfig) GetUserEligibilitySchedules(subscriptionId string) []*Schedule {
	return getResourcesEligibilitySchedules(c.Users, subscriptionId)
}

func (c *ResourcesConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterStructValidation(ResourcesConfigStructLevelValidation, ResourcesConfig{})
	validate.RegisterStructValidation(ResourceConfigurationStructLevelValidation, ResourceConfiguration{})

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func ResourcesConfigStructLevelValidation(sl validator.StructLevel) {
	ResourcesConfig := sl.Current().Interface().(ResourcesConfig)

	var rulesetReferences []*RulesetReference
	for _, p := range ResourcesConfig.Policies {
		rulesetReferences = append(rulesetReferences, p.Default...)
		rulesetReferences = append(rulesetReferences, p.Subscription...)
		for _, r := range p.ResourceGroups {
			rulesetReferences = append(rulesetReferences, r...)
		}
		for _, r := range p.Resources {
			rulesetReferences = append(rulesetReferences, r...)
		}
	}

	for _, r := range rulesetReferences {
		any := linq.From(ResourcesConfig.Rulesets).WhereT(func(s *RoleManagementPolicyRuleset) bool {
			return s.Name == r.RulesetName
		}).Any()
		if !any {
			sl.ReportError(r.RulesetName, "Rulesets", "", fmt.Sprintf("ruleset %s not found", r.RulesetName), "")
		}
	}

	// TODO: Check for policy conflicts.
}

func ResourceConfigurationStructLevelValidation(sl validator.StructLevel) {
	resourceConfiguration := sl.Current().Interface().(ResourceConfiguration)

	if countUniqueSchedules(resourceConfiguration.Active) != len(resourceConfiguration.Active) {
		sl.ReportError(resourceConfiguration.Active, "Active", "", "duplicate active role name", "")
	}

	if countUniqueSchedules(resourceConfiguration.Eligible) != len(resourceConfiguration.Eligible) {
		sl.ReportError(resourceConfiguration.Eligible, "Eligible", "", "duplicate eligible role name", "")
	}
}

func countUniqueSchedules(schedules []*Schedule) int {
	seen := make(map[string]bool)
	unique := []string{}

	for _, s := range schedules {
		if _, ok := seen[s.RoleName]; !ok {
			seen[s.RoleName] = true
			unique = append(unique, s.RoleName)
		}
	}
	return len(unique)
}

func getResourcesAssignmentSchedules(principals []*ResourcePrincipal, subscriptionId string) []*Schedule {
	schedules := []*Schedule{}

	for _, p := range principals {
		if p.Subscription != nil {
			for _, s := range p.Subscription.Active {
				s.PrincipalName = p.Name
				s.Target = fmt.Sprintf("/subscriptions/%s", subscriptionId)
				schedules = append(schedules, s)
			}
		}

		resourceGroupNames := make([]string, 0, len(p.ResourceGroups))
		for k := range p.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, s := range p.ResourceGroups[r].Active {
				s.PrincipalName = p.Name
				s.Target = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				schedules = append(schedules, s)
			}
		}

		resourceNames := make([]string, 0, len(p.Resources))
		for k := range p.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, s := range p.Resources[r].Active {
				s.PrincipalName = p.Name
				s.Target = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				schedules = append(schedules, s)
			}
		}
	}

	return schedules
}

func getResourcesEligibilitySchedules(principals []*ResourcePrincipal, subscriptionId string) []*Schedule {
	schedules := []*Schedule{}

	for _, p := range principals {
		if p.Subscription != nil {
			for _, s := range p.Subscription.Eligible {
				s.PrincipalName = p.Name
				s.Target = fmt.Sprintf("/subscriptions/%s", subscriptionId)
				schedules = append(schedules, s)
			}
		}

		resourceGroupNames := make([]string, 0, len(p.ResourceGroups))
		for k := range p.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, s := range p.ResourceGroups[r].Eligible {
				s.PrincipalName = p.Name
				s.Target = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				schedules = append(schedules, s)
			}
		}

		resourceNames := make([]string, 0, len(p.Resources))
		for k := range p.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, s := range p.Resources[r].Eligible {
				s.PrincipalName = p.Name
				s.Target = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				schedules = append(schedules, s)
			}
		}
	}

	return schedules
}
