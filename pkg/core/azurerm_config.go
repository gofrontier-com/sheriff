package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func (c *AzureRmConfig) GetGroupAssignmentSchedules(subscriptionId string) []*Schedule {
	return getAssignmentSchedules(c.Groups, subscriptionId)
}

func (c *AzureRmConfig) GetGroupEligibilitySchedules(subscriptionId string) []*Schedule {
	return getEligibilitySchedules(c.Groups, subscriptionId)
}

func (c *AzureRmConfig) GetUserAssignmentSchedules(subscriptionId string) []*Schedule {
	return getAssignmentSchedules(c.Users, subscriptionId)
}

func (c *AzureRmConfig) GetUserEligibilitySchedules(subscriptionId string) []*Schedule {
	return getEligibilitySchedules(c.Users, subscriptionId)
}

func (c *AzureRmConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	// validate.RegisterStructValidation(ActiveAssignmentsStructLevelValidation, activeAssignments{})
	validate.RegisterStructValidation(AzureRmConfigStructLevelValidation, AzureRmConfig{})

	// TODO: Catch missing policies
	// TODO: Validate eligible assignments
	// TODO: Validate policies and policy conflicts

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

// func ActiveAssignmentsStructLevelValidation(sl validator.StructLevel) {
// 	activeAssignments := sl.Current().Interface().(activeAssignments)

// 	if CountUniqueActiveAssignments(activeAssignments.Subscription) != len(activeAssignments.Subscription) {
// 		sl.ReportError(activeAssignments.Subscription, "Subscription", "", "duplicate role name", "")
// 	}

// 	resourceGroupNames := make([]string, 0, len(activeAssignments.ResourceGroups))
// 	for g := range activeAssignments.ResourceGroups {
// 		resourceGroupNames = append(resourceGroupNames, g)
// 	}

// 	for _, r := range resourceGroupNames {
// 		if CountUniqueActiveAssignments(activeAssignments.ResourceGroups[r]) != len(activeAssignments.ResourceGroups[r]) {
// 			sl.ReportError(activeAssignments.ResourceGroups[r], r, "", "duplicate role name", "")
// 		}
// 	}

// 	resourceNames := make([]string, 0, len(activeAssignments.Resources))
// 	for r := range activeAssignments.Resources {
// 		resourceNames = append(resourceNames, r)
// 	}

// 	for _, r := range resourceNames {
// 		if CountUniqueActiveAssignments(activeAssignments.Resources[r]) != len(activeAssignments.Resources[r]) {
// 			sl.ReportError(activeAssignments.Resources[r], r, "", "duplicate role name", "")
// 		}
// 	}
// }

func AzureRmConfigStructLevelValidation(sl validator.StructLevel) {
	azureRmConfig := sl.Current().Interface().(AzureRmConfig)
	_ = azureRmConfig
}

// func CountUniqueActiveAssignments(activeAssignments []*ActiveAssignment) int {
// 	seen := make(map[string]bool)
// 	unique := []string{}

// 	for _, a := range activeAssignments {
// 		if _, ok := seen[a.RoleName]; !ok {
// 			seen[a.RoleName] = true
// 			unique = append(unique, a.RoleName)
// 		}
// 	}
// 	return len(unique)
// }

func getAssignmentSchedules(principals []*Principal, subscriptionId string) []*Schedule {
	assignmentSchedules := []*Schedule{}

	for _, p := range principals {
		if p.Subscription == nil {
			continue
		}

		for _, e := range p.Subscription.Active {
			e.PrincipalName = p.Name
			e.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
			assignmentSchedules = append(assignmentSchedules, e)
		}

		resourceGroupNames := make([]string, 0, len(p.ResourceGroups))
		for k := range p.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, e := range p.ResourceGroups[r].Active {
				e.PrincipalName = p.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				assignmentSchedules = append(assignmentSchedules, e)
			}
		}

		resourceNames := make([]string, 0, len(p.Resources))
		for k := range p.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, e := range p.Resources[r].Active {
				e.PrincipalName = p.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				assignmentSchedules = append(assignmentSchedules, e)
			}
		}
	}

	return assignmentSchedules
}

func getEligibilitySchedules(principals []*Principal, subscriptionId string) []*Schedule {
	eligibilitySchedules := []*Schedule{}

	for _, p := range principals {
		if p.Subscription == nil {
			continue
		}

		for _, e := range p.Subscription.Eligible {
			e.PrincipalName = p.Name
			e.Scope = fmt.Sprintf("/subscriptions/%s", subscriptionId)
			eligibilitySchedules = append(eligibilitySchedules, e)
		}

		resourceGroupNames := make([]string, 0, len(p.ResourceGroups))
		for k := range p.ResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}

		for _, r := range resourceGroupNames {
			for _, e := range p.ResourceGroups[r].Eligible {
				e.PrincipalName = p.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				eligibilitySchedules = append(eligibilitySchedules, e)
			}
		}

		resourceNames := make([]string, 0, len(p.Resources))
		for k := range p.Resources {
			resourceNames = append(resourceNames, k)
		}

		for _, r := range resourceNames {
			for _, e := range p.Resources[r].Eligible {
				e.PrincipalName = p.Name
				e.Scope = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, r)
				eligibilitySchedules = append(eligibilitySchedules, e)
			}
		}
	}

	return eligibilitySchedules
}
