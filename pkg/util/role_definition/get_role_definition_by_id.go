package role_definition

import (
	"fmt"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleDefinitionById(clientFactory *armauthorization.ClientFactory, cache gocache.Cache, subscriptionId string, roleDefinitionId string) (*armauthorization.RoleDefinition, error) {
	roleDefinitions, err := GetRoleDefinitions(clientFactory, cache, subscriptionId)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(roleDefinitions, func(r *armauthorization.RoleDefinition) bool {
		return *r.ID == roleDefinitionId
	})

	if idx == -1 {
		return nil, fmt.Errorf("role definition with Id \"%s\" not found", roleDefinitionId)
	}

	return roleDefinitions[idx], nil
}
