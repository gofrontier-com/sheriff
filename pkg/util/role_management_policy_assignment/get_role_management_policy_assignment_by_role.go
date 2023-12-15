package role_management_policy_assignment

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func GetRoleManagementPolicyAssignmentByRole(clientFactory *armauthorization.ClientFactory, scope string, roleName string) (*armauthorization.RoleManagementPolicyAssignment, error) {
	roleManagementPolicyAssignments, err := GetRoleManagementPolicyAssignments(
		clientFactory,
		scope,
		func(r *armauthorization.RoleManagementPolicyAssignment) bool {
			return *r.Properties.PolicyAssignmentProperties.RoleDefinition.DisplayName == roleName
		},
	)
	if err != nil {
		return nil, err
	}

	if len(roleManagementPolicyAssignments) == 0 {
		return nil, fmt.Errorf("role management policy assignment for role name \"%s\" not found", roleName)
	}

	if len(roleManagementPolicyAssignments) > 1 {
		return nil, fmt.Errorf("multiple role management policy assignment for role name \"%s\" found", roleName)
	}

	return roleManagementPolicyAssignments[0], nil
}
