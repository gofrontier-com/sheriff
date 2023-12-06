package role_assignment

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleAssignments(clientFactory *armauthorization.ClientFactory, cache gocache.Cache, scope string, filter func(*armauthorization.RoleAssignment) bool) ([]*armauthorization.RoleAssignment, error) {
	var roleAssignments []*armauthorization.RoleAssignment
	cacheKey := "roleAssignments"

	if r, found := cache.Get(cacheKey); found {
		roleAssignments = r.([]*armauthorization.RoleAssignment)
	} else {
		roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()

		pager := roleAssignmentsClient.NewListForScopePager(scope, nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			for _, r := range page.Value {
				if strings.HasPrefix(*r.Properties.Scope, scope) {
					roleAssignments = append(roleAssignments, r)
				}
			}
		}

		cache.Set(cacheKey, roleAssignments, gocache.NoExpiration)
	}

	var filteredRoleAssignments []*armauthorization.RoleAssignment
	for _, r := range roleAssignments {
		if filter(r) {
			filteredRoleAssignments = append(filteredRoleAssignments, r)
		}
	}

	return filteredRoleAssignments, nil
}
