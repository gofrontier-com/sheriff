package role_definition

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleDefinitions(clientFactory *armauthorization.ClientFactory, cache gocache.Cache, subscriptionId string) ([]*armauthorization.RoleDefinition, error) {
	var roleDefinitions []*armauthorization.RoleDefinition
	cacheKey := "roleDefinitions"

	if r, found := cache.Get(cacheKey); found {
		roleDefinitions = r.([]*armauthorization.RoleDefinition)
	} else {
		roleDefinitionsClient := clientFactory.NewRoleDefinitionsClient()

		pager := roleDefinitionsClient.NewListPager(fmt.Sprintf("/subscriptions/%s", subscriptionId), nil)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			roleDefinitions = append(roleDefinitions, page.Value...)
		}

		cache.Set(cacheKey, roleDefinitions, gocache.NoExpiration)
	}

	return roleDefinitions, nil
}
