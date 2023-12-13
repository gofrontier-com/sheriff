package role_management_policy_assignment

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleManagementPolicyAssignmentByRole(clientFactory *armauthorization.ClientFactory, cache gocache.Cache, scope string, roleName string) {

}
