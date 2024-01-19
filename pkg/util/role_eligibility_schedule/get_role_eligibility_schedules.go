package role_eligibility_schedule

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

func GetRoleEligibilitySchedules(clientFactory *armauthorization.ClientFactory, scope string, filter func(*armauthorization.RoleEligibilitySchedule) bool) ([]*armauthorization.RoleEligibilitySchedule, error) {
	var roleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule
	cacheKey := fmt.Sprintf("roleEligibilitySchedules_%s", scope)

	if s, found := cache.Get(cacheKey); found {
		roleEligibilitySchedules = s.([]*armauthorization.RoleEligibilitySchedule)
	} else {
		roleEligibilitySchedulesClient := clientFactory.NewRoleEligibilitySchedulesClient()

		pager := roleEligibilitySchedulesClient.NewListForScopePager(scope, nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			for _, s := range page.Value {
				if strings.HasPrefix(*s.Properties.Scope, scope) {
					roleEligibilitySchedules = append(roleEligibilitySchedules, s)
				}
			}
		}

		cache.Set(cacheKey, roleEligibilitySchedules, gocache.NoExpiration)
	}

	var filteredRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule
	for _, s := range roleEligibilitySchedules {
		if filter(s) {
			filteredRoleEligibilitySchedules = append(filteredRoleEligibilitySchedules, s)
		}
	}

	return filteredRoleEligibilitySchedules, nil
}
