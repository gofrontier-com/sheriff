package role_assignment

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleAssignments(roleAssignmentsClient *armauthorization.RoleAssignmentsClient, cache gocache.Cache, subscriptionId string, filter func(*armauthorization.RoleAssignment) bool) ([]*armauthorization.RoleAssignment, error) {
	var roleAssignments []*armauthorization.RoleAssignment
	cacheKey := "roleAssignments"

	if r, found := cache.Get(cacheKey); found {
		roleAssignments = r.([]*armauthorization.RoleAssignment)
	} else {
		pager := roleAssignmentsClient.NewListForSubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			for _, r := range page.Value {
				if strings.HasPrefix(*r.Properties.Scope, fmt.Sprintf("/subscriptions/%s", subscriptionId)) {
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
