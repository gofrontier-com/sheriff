package core

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/go-playground/validator/v10"
)

// TODO: Validate.

func (c *GroupsConfig) GetGroupAssignmentSchedules() []*Schedule {
	return getGroupsAssignmentSchedules(c.Groups)
}

func (c *GroupsConfig) GetGroupEligibilitySchedules() []*Schedule {
	return getGroupsEligibilitySchedules(c.Groups)
}

func (c *GroupsConfig) GetGroupNameRoleNameCombinations() []*TargetRoleNameCombination {
	groupAssignmentSchedules := c.GetGroupAssignmentSchedules()
	userAssignmentSchedules := c.GetUserAssignmentSchedules()
	groupEligibilitySchedules := c.GetGroupEligibilitySchedules()
	userEligibilitySchedules := c.GetUserEligibilitySchedules()

	allSchedules := append(groupAssignmentSchedules, userAssignmentSchedules...)
	allSchedules = append(allSchedules, groupEligibilitySchedules...)
	allSchedules = append(allSchedules, userEligibilitySchedules...)

	var groupNameRoleNameCombinations []*TargetRoleNameCombination
	linq.From(allSchedules).SelectT(func(s *Schedule) *TargetRoleNameCombination {
		return &TargetRoleNameCombination{
			RoleName: s.RoleName,
			Target:   s.Target,
		}
	}).DistinctByT(func(s *TargetRoleNameCombination) string {
		return fmt.Sprintf("%s:%s", s.Target, s.RoleName)
	}).ToSlice(&groupNameRoleNameCombinations)

	return groupNameRoleNameCombinations
}

func (c *GroupsConfig) GetPolicyByRoleName(roleName string) *GroupPolicy {
	policy := linq.From(c.Policies).SingleWithT(func(p *GroupPolicy) bool {
		return p.Name == roleName
	})

	if policy == nil {
		policy = linq.From(c.Policies).SingleWithT(func(p *GroupPolicy) bool {
			return p.Name == "default"
		})
	}

	if p, ok := policy.(*GroupPolicy); ok {
		return p
	} else {
		return nil
	}
}

func (c *GroupsConfig) GetUserAssignmentSchedules() []*Schedule {
	return getGroupsAssignmentSchedules(c.Users)
}

func (c *GroupsConfig) GetUserEligibilitySchedules() []*Schedule {
	return getGroupsEligibilitySchedules(c.Users)
}

func (c *GroupsConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	// validate.RegisterStructValidation(GroupsConfigStructLevelValidation, GroupsConfig{})
	validate.RegisterStructValidation(ResourceConfigurationStructLevelValidation, ResourceConfiguration{})

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func getGroupsAssignmentSchedules(principals []*GroupPrincipal) []*Schedule {
	schedules := []*Schedule{}

	for _, p := range principals {
		managedGroupNames := make([]string, 0, len(p.ManagedGroups))
		for k := range p.ManagedGroups {
			managedGroupNames = append(managedGroupNames, k)
		}

		for _, m := range managedGroupNames {
			for _, s := range p.ManagedGroups[m].Active {
				s.PrincipalName = p.Name
				s.Target = m
				schedules = append(schedules, s)
			}
		}
	}

	return schedules
}

func getGroupsEligibilitySchedules(principals []*GroupPrincipal) []*Schedule {
	schedules := []*Schedule{}

	for _, p := range principals {
		managedGroupNames := make([]string, 0, len(p.ManagedGroups))
		for k := range p.ManagedGroups {
			managedGroupNames = append(managedGroupNames, k)
		}

		for _, m := range managedGroupNames {
			for _, s := range p.ManagedGroups[m].Eligible {
				s.PrincipalName = p.Name
				s.Target = m
				schedules = append(schedules, s)
			}
		}
	}

	return schedules
}
