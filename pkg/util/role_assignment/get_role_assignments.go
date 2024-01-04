package role_assignment

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func GetRoleAssignments(clientFactory *armauthorization.ClientFactory, scope string, filter func(*armauthorization.RoleAssignment) bool) ([]*armauthorization.RoleAssignment, error) {
	var roleAssignments []*armauthorization.RoleAssignment
	cacheKey := fmt.Sprintf("roleAssignments_%s", scope)

	if a, found := cache.Get(cacheKey); found {
		roleAssignments = a.([]*armauthorization.RoleAssignment)
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
	for _, a := range roleAssignments {
		if filter(a) {
			filteredRoleAssignments = append(filteredRoleAssignments, a)
		}
	}

	return filteredRoleAssignments, nil
}
