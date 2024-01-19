package role_management_policy

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func GetRoleManagementPolicyById(clientFactory *armauthorization.ClientFactory, scope string, roleManagementPolicyId string) (*armauthorization.RoleManagementPolicy, error) {
	roleManagementPoliciesClient := clientFactory.NewRoleManagementPoliciesClient()

	response, err := roleManagementPoliciesClient.Get(context.Background(), scope, roleManagementPolicyId, nil)
	if err != nil {
		return nil, err
	}

	// There's a bug in the SDK where the response is not nil even if you pass a bad Id.
	if response.RoleManagementPolicy.ID == nil {
		return nil, fmt.Errorf("role management policy with Id '%s' not found", roleManagementPolicyId)
	}

	return &response.RoleManagementPolicy, err
}
