package core

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/go-playground/validator/v10"
)

func (c *AzureRmConfig) GetGroupAssignmentSchedules(subscriptionId string) []*Schedule {
	return getAssignmentSchedules(c.Groups, subscriptionId)
}

func (c *AzureRmConfig) GetGroupEligibilitySchedules(subscriptionId string) []*Schedule {
	return getEligibilitySchedules(c.Groups, subscriptionId)
}

func (c *AzureRmConfig) GetRulesetReferences(subscriptionId string) []*RulesetReference {
	rulesetReferences := []*RulesetReference{}

	for _, p := range c.Policies {
		if p.Default != nil {
			for _, r := range p.Default {
				r.RoleName = p.Name
				r.Scope = "default"
				rulesetReferences = append(rulesetReferences, r)
			}
		}

		if p.Subscription != nil {
			for _, r := range p.Subscription {
				r.RoleName = p.Name
				r.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
				rulesetReferences = append(rulesetReferences, r)
			}
		}

		resourceGroupNames := make([]string, 0, len(p.ResourceGroups))
		for k := range p.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, s := range p.ResourceGroups[r] {
				s.RoleName = p.Name
				s.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				rulesetReferences = append(rulesetReferences, s)
			}
		}

		resourceNames := make([]string, 0, len(p.Resources))
		for k := range p.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, s := range p.Resources[r] {
				s.RoleName = p.Name
				s.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				rulesetReferences = append(rulesetReferences, s)
			}
		}
	}

	return rulesetReferences
}

func (c *AzureRmConfig) GetScopeRoleNameCombinations(subscriptionId string) []*ScopeRoleNameCombination {
	groupAssignmentSchedules := c.GetGroupAssignmentSchedules(subscriptionId)
	userAssignmentSchedules := c.GetUserAssignmentSchedules(subscriptionId)
	groupEligibilitySchedules := c.GetGroupEligibilitySchedules(subscriptionId)
	userEligibilitySchedules := c.GetUserEligibilitySchedules(subscriptionId)

	allSchedules := append(groupAssignmentSchedules, userAssignmentSchedules...)
	allSchedules = append(allSchedules, groupEligibilitySchedules...)
	allSchedules = append(allSchedules, userEligibilitySchedules...)

	var scopeRoleNameCombinations []*ScopeRoleNameCombination
	linq.From(allSchedules).SelectT(func(s *Schedule) *ScopeRoleNameCombination {
		return &ScopeRoleNameCombination{
			RoleName: s.RoleName,
			Scope:    s.Scope,
		}
	}).DistinctByT(func(s *ScopeRoleNameCombination) string {
		return fmt.Sprintf("%s:%s", s.Scope, s.RoleName)
	}).ToSlice(&scopeRoleNameCombinations)

	return scopeRoleNameCombinations
}

func (c *AzureRmConfig) GetUserAssignmentSchedules(subscriptionId string) []*Schedule {
	return getAssignmentSchedules(c.Users, subscriptionId)
}

func (c *AzureRmConfig) GetUserEligibilitySchedules(subscriptionId string) []*Schedule {
	return getEligibilitySchedules(c.Users, subscriptionId)
}

func (c *AzureRmConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterStructValidation(AzureRmConfigStructLevelValidation, AzureRmConfig{})
	validate.RegisterStructValidation(ScopeConfigurationStructLevelValidation, ScopeConfiguration{})

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func AzureRmConfigStructLevelValidation(sl validator.StructLevel) {
	azureRmConfig := sl.Current().Interface().(AzureRmConfig)

	rulesetReferences := azureRmConfig.GetRulesetReferences("00000000-0000-0000-0000-000000000000")

	for _, r := range rulesetReferences {
		any := linq.From(azureRmConfig.Rulesets).WhereT(func(s *RoleManagementPolicyRuleset) bool {
			return s.Name == r.RulesetName
		}).Any()
		if !any {
			sl.ReportError(r.RulesetName, "Rulesets", "", fmt.Sprintf("ruleset %s not found", r.RulesetName), "")
		}
	}

	// TODO: Check for policy conflicts.
}

func ScopeConfigurationStructLevelValidation(sl validator.StructLevel) {
	scopeConfiguration := sl.Current().Interface().(ScopeConfiguration)

	if countUniqueSchedules(scopeConfiguration.Active) != len(scopeConfiguration.Active) {
		sl.ReportError(scopeConfiguration.Active, "Active", "", "duplicate active role name", "")
	}

	if countUniqueSchedules(scopeConfiguration.Eligible) != len(scopeConfiguration.Eligible) {
		sl.ReportError(scopeConfiguration.Eligible, "Eligible", "", "duplicate eligible role name", "")
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

func getAssignmentSchedules(principals []*Principal, subscriptionId string) []*Schedule {
	schedules := []*Schedule{}

	for _, p := range principals {
		if p.Subscription != nil {
			for _, s := range p.Subscription.Active {
				s.PrincipalName = p.Name
				s.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
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
				s.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
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
				s.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				schedules = append(schedules, s)
			}
		}
	}

	return schedules
}

func getEligibilitySchedules(principals []*Principal, subscriptionId string) []*Schedule {
	schedules := []*Schedule{}

	for _, p := range principals {
		if p.Subscription != nil {
			for _, s := range p.Subscription.Eligible {
				s.PrincipalName = p.Name
				s.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
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
				s.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
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
				s.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				schedules = append(schedules, s)
			}
		}
	}

	return schedules
}
