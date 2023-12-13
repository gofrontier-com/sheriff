package role_management_policy_assignment

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func GetRoleManagementPolicyAssignments(clientFactory *armauthorization.ClientFactory, scope string, filter func(*armauthorization.RoleManagementPolicyAssignment) bool) ([]*armauthorization.RoleManagementPolicyAssignment, error) {
	var roleManagementPolicyAssignments []*armauthorization.RoleManagementPolicyAssignment
	cacheKey := fmt.Sprintf("roleManagementPolicyAssignments_%s", scope)

	if a, found := cache.Get(cacheKey); found {
		roleManagementPolicyAssignments = a.([]*armauthorization.RoleManagementPolicyAssignment)
	} else {
		roleManagementPolicyAssignmentsClient := clientFactory.NewRoleManagementPolicyAssignmentsClient()

		pager := roleManagementPolicyAssignmentsClient.NewListForScopePager(scope, nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			roleManagementPolicyAssignments = append(roleManagementPolicyAssignments, page.Value...)
		}

		cache.Set(cacheKey, roleManagementPolicyAssignments, gocache.NoExpiration)
	}

	var filteredRoleManagementPolicyAssignments []*armauthorization.RoleManagementPolicyAssignment
	for _, s := range roleManagementPolicyAssignments {
		if filter(s) {
			filteredRoleManagementPolicyAssignments = append(filteredRoleManagementPolicyAssignments, s)
		}
	}

	return filteredRoleManagementPolicyAssignments, nil
}
