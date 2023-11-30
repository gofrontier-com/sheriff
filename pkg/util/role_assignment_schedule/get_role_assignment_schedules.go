package role_assignment_schedule

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleAssignmentSchedules(roleAssignmentSchedulesClient *armauthorization.RoleAssignmentSchedulesClient, cache gocache.Cache, subscriptionId string, filter func(*armauthorization.RoleAssignmentSchedule) bool) ([]*armauthorization.RoleAssignmentSchedule, error) {
	var roleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule
	cacheKey := "roleAssignmentSchedules"

	if r, found := cache.Get(cacheKey); found {
		roleAssignmentSchedules = r.([]*armauthorization.RoleAssignmentSchedule)
	} else {
		pager := roleAssignmentSchedulesClient.NewListForScopePager(fmt.Sprintf("/subscriptions/%s", subscriptionId), nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			// for _, r := range page.Value {
			// 	if strings.HasPrefix(*r.Properties.Scope, fmt.Sprintf("/subscriptions/%s", subscriptionId)) {
			// 		roleAssignments = append(roleAssignments, r)
			// 	}
			// }

			roleAssignmentSchedules = append(roleAssignmentSchedules, page.Value...)
		}

		cache.Set(cacheKey, roleAssignmentSchedules, gocache.NoExpiration)
	}

	var filteredRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule
	for _, r := range roleAssignmentSchedules {
		if filter(r) {
			filteredRoleAssignmentSchedules = append(filteredRoleAssignmentSchedules, r)
		}
	}

	return filteredRoleAssignmentSchedules, nil
}
