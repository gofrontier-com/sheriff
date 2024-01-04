package role_definition

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

func GetRoleDefinitions(clientFactory *armauthorization.ClientFactory, scope string, filter func(*armauthorization.RoleDefinition) bool) ([]*armauthorization.RoleDefinition, error) {
	var roleDefinitions []*armauthorization.RoleDefinition
	cacheKey := fmt.Sprintf("roleDefinitions_%s", scope)

	if d, found := cache.Get(cacheKey); found {
		roleDefinitions = d.([]*armauthorization.RoleDefinition)
	} else {
		roleDefinitionsClient := clientFactory.NewRoleDefinitionsClient()

		pager := roleDefinitionsClient.NewListPager(scope, nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			roleDefinitions = append(roleDefinitions, page.Value...)
		}

		cache.Set(cacheKey, roleDefinitions, gocache.NoExpiration)
	}

	var filteredRoleDefinitions []*armauthorization.RoleDefinition
	for _, a := range roleDefinitions {
		if filter(a) {
			filteredRoleDefinitions = append(filteredRoleDefinitions, a)
		}
	}

	return filteredRoleDefinitions, nil
}
