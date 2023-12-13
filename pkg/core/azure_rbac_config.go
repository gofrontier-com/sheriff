package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func (c *AzureRmConfig) GetGroupActiveAssignments(subscriptionId string) []*ActiveAssignment {
	return getActiveAssignments(c.Groups, subscriptionId)
}

func (c *AzureRmConfig) GetGroupEligibleAssignments(subscriptionId string) []*EligibleAssignment {
	return getEligibleAssignments(c.Groups, subscriptionId)
}

// func (c *AzureRmConfig) GetRoleManagementPolicyRulesets() []*RoleManagementPolicyRuleset {
// 	groupEligibleAssignments := c.GetGroupEligibleAssignments("")
// }

func (c *AzureRmConfig) GetUserActiveAssignments(subscriptionId string) []*ActiveAssignment {
	return getActiveAssignments(c.Users, subscriptionId)
}

func (c *AzureRmConfig) GetUserEligibleAssignments(subscriptionId string) []*EligibleAssignment {
	return getEligibleAssignments(c.Users, subscriptionId)
}

func (c *AzureRmConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func getActiveAssignments(principals []*Principal, subscriptionId string) []*ActiveAssignment {
	activeAssignments := []*ActiveAssignment{}

	for _, p := range principals {
		if p.Active == nil {
			continue
		}

		for _, a := range p.Active.Subscription {
			a.PrincipalName = p.Name
			a.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
			activeAssignments = append(activeAssignments, a)
		}

		resourceGroupNames := make([]string, 0, len(p.Active.ResourceGroups))
		for k := range p.Active.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, a := range p.Active.ResourceGroups[r] {
				a.PrincipalName = p.Name
				a.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				activeAssignments = append(activeAssignments, a)
			}
		}

		resourceNames := make([]string, 0, len(p.Active.Resources))
		for k := range p.Active.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, a := range p.Active.Resources[r] {
				a.PrincipalName = p.Name
				a.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				activeAssignments = append(activeAssignments, a)
			}
		}
	}

	return activeAssignments
}

func getEligibleAssignments(principals []*Principal, subscriptionId string) []*EligibleAssignment {
	eligibleAssignments := []*EligibleAssignment{}

	for _, p := range principals {
		if p.Eligible == nil {
			continue
		}

		for _, e := range p.Eligible.Subscription {
			e.PrincipalName = p.Name
			e.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
			eligibleAssignments = append(eligibleAssignments, e)
		}

		resourceGroupNames := make([]string, 0, len(p.Eligible.ResourceGroups))
		for k := range p.Eligible.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, e := range p.Eligible.ResourceGroups[r] {
				e.PrincipalName = p.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				eligibleAssignments = append(eligibleAssignments, e)
			}
		}

		resourceNames := make([]string, 0, len(p.Eligible.Resources))
		for k := range p.Eligible.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, e := range p.Eligible.Resources[r] {
				e.PrincipalName = p.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				eligibleAssignments = append(eligibleAssignments, e)
			}
		}
	}

	return eligibleAssignments
}
