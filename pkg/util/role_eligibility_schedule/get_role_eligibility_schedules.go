package role_eligibility_schedule

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleEligibilitySchedules(clientFactory *armauthorization.ClientFactory, cache gocache.Cache, scope string, filter func(*armauthorization.RoleEligibilitySchedule) bool) ([]*armauthorization.RoleEligibilitySchedule, error) {
	var roleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule
	cacheKey := "roleEligibilitySchedules"

	if r, found := cache.Get(cacheKey); found {
		roleEligibilitySchedules = r.([]*armauthorization.RoleEligibilitySchedule)
	} else {
		roleEligibilitySchedulesClient := clientFactory.NewRoleEligibilitySchedulesClient()

		pager := roleEligibilitySchedulesClient.NewListForScopePager(scope, nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			for _, r := range page.Value {
				if strings.HasPrefix(*r.Properties.Scope, scope) {
					roleEligibilitySchedules = append(roleEligibilitySchedules, r)
				}
			}
		}

		cache.Set(cacheKey, roleEligibilitySchedules, gocache.NoExpiration)
	}

	var filteredRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule
	for _, r := range roleEligibilitySchedules {
		if filter(r) {
			filteredRoleEligibilitySchedules = append(filteredRoleEligibilitySchedules, r)
		}
	}

	return filteredRoleEligibilitySchedules, nil
}