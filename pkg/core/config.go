package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func (c *Config) GetActiveAssignments(subscriptionId string) []*ActiveAssignment {
	activeAssignments := []*ActiveAssignment{}

	for _, g := range c.Groups {
		if g.Active == nil {
			continue
		}

		for _, a := range g.Active.Subscription {
			a.GroupName = g.Name
			a.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
			activeAssignments = append(activeAssignments, a)
		}

		resourceGroupNames := make([]string, 0, len(g.Active.ResourceGroups))
		for k := range g.Active.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, a := range g.Active.ResourceGroups[r] {
				a.GroupName = g.Name
				a.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				activeAssignments = append(activeAssignments, a)
			}
		}

		resourceNames := make([]string, 0, len(g.Active.Resources))
		for k := range g.Active.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, a := range g.Active.Resources[r] {
				a.GroupName = g.Name
				a.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				activeAssignments = append(activeAssignments, a)
			}
		}
	}

	return activeAssignments
}

func (c *Config) GetEligibleAssignments(subscriptionId string) []*EligibleAssignment {
	eligibleAssignments := []*EligibleAssignment{}

	for _, g := range c.Groups {
		if g.Eligible == nil {
			continue
		}

		for _, e := range g.Eligible.Subscription {
			e.GroupName = g.Name
			e.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
			eligibleAssignments = append(eligibleAssignments, e)
		}

		resourceGroupNames := make([]string, 0, len(g.Eligible.ResourceGroups))
		for k := range g.Eligible.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, e := range g.Eligible.ResourceGroups[r] {
				e.GroupName = g.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				eligibleAssignments = append(eligibleAssignments, e)
			}
		}

		resourceNames := make([]string, 0, len(g.Eligible.Resources))
		for k := range g.Eligible.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, e := range g.Eligible.Resources[r] {
				e.GroupName = g.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				eligibleAssignments = append(eligibleAssignments, e)
			}
		}
	}

	return eligibleAssignments
}

func (c *Config) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}
