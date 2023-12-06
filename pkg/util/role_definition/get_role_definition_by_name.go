package role_definition

import (
	"fmt"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleDefinitionByName(clientFactory *armauthorization.ClientFactory, cache gocache.Cache, scope string, roleDefinitionName string) (*armauthorization.RoleDefinition, error) {
	roleDefinitions, err := GetRoleDefinitions(clientFactory, cache, scope)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(roleDefinitions, func(r *armauthorization.RoleDefinition) bool {
		return *r.Properties.RoleName == roleDefinitionName
	})

	if idx == -1 {
		return nil, fmt.Errorf("role definition with name \"%s\" not found", roleDefinitionName)
	}

	return roleDefinitions[idx], nil
}
