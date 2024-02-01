package role_assignment_schedule

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

func GetRoleAssignmentSchedules(clientFactory *armauthorization.ClientFactory, scope string, filter func(*armauthorization.RoleAssignmentSchedule) bool) ([]*armauthorization.RoleAssignmentSchedule, error) {
	var roleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule
	cacheKey := fmt.Sprintf("roleAssignmentSchedules_%s", scope)

	if s, found := cache.Get(cacheKey); found {
		roleAssignmentSchedules = s.([]*armauthorization.RoleAssignmentSchedule)
	} else {
		roleAssignmentSchedulesClient := clientFactory.NewRoleAssignmentSchedulesClient()

		pager := roleAssignmentSchedulesClient.NewListForScopePager(scope, nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			for _, s := range page.Value {
				if strings.HasPrefix(*s.Properties.Scope, scope) {
					roleAssignmentSchedules = append(roleAssignmentSchedules, s)
				}
			}
		}

		cache.Set(cacheKey, roleAssignmentSchedules, gocache.NoExpiration)
	}

	var filteredRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule
	for _, s := range roleAssignmentSchedules {
		if filter(s) {
			filteredRoleAssignmentSchedules = append(filteredRoleAssignmentSchedules, s)
		}
	}

	return filteredRoleAssignmentSchedules, nil
}
