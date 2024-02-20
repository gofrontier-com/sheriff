package role_definition

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleDefinitionByName(clientFactory *armauthorization.ClientFactory, scope string, roleDefinitionName string) (*armauthorization.RoleDefinition, error) {
	var roleDefinition *armauthorization.RoleDefinition
	cacheKey := fmt.Sprintf("scoped-name::%s_%s", scope, roleDefinitionName)

	if d, found := cache.Get(cacheKey); found {
		roleDefinition = d.(*armauthorization.RoleDefinition)
	} else {
		roleDefinitionsClient := clientFactory.NewRoleDefinitionsClient()

		var roleDefinitions []*armauthorization.RoleDefinition
		options := &armauthorization.RoleDefinitionsClientListOptions{
			Filter: to.Ptr(fmt.Sprintf("roleName eq '%s'", roleDefinitionName)),
		}
		pager := roleDefinitionsClient.NewListPager(scope, options)
		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			roleDefinitions = append(roleDefinitions, page.Value...)
		}

		if len(roleDefinitions) == 0 {
			return nil, fmt.Errorf("role definition with name \"%s\" not found", roleDefinitionName)
		}

		if len(roleDefinitions) > 1 {
			return nil, fmt.Errorf("multiple role definition with name \"%s\" found", roleDefinitionName)
		}

		cache.Set(cacheKey, roleDefinitions[0], gocache.NoExpiration)

		roleDefinition = roleDefinitions[0]
	}

	return roleDefinition, nil
}
