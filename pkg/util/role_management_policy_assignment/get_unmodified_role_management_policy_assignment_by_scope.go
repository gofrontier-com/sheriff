package role_management_policy_assignment

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func GetUnmodifiedRoleManagementPolicyAssignmentByScope(clientFactory *armauthorization.ClientFactory, scope string) (*armauthorization.RoleManagementPolicyAssignment, error) {
	roleManagementPolicyAssignments, err := GetRoleManagementPolicyAssignments(
		clientFactory,
		scope,
		func(r *armauthorization.RoleManagementPolicyAssignment) bool {
			return r.Properties.PolicyAssignmentProperties.Policy.LastModifiedDateTime == nil
		},
	)
	if err != nil {
		return nil, err
	}

	if len(roleManagementPolicyAssignments) == 0 {
		return nil, fmt.Errorf("no unmodified role management policy assignments at scope \"%s\" found", scope)
	}

	return roleManagementPolicyAssignments[0], nil
}
