package role_definition

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func GetRoleDefinitionById(clientFactory *armauthorization.ClientFactory, scope string, roleDefinitionId string) (*armauthorization.RoleDefinition, error) {
	roleDefinitions, err := GetRoleDefinitions(
		clientFactory,
		scope,
		func(r *armauthorization.RoleDefinition) bool {
			return *r.ID == roleDefinitionId
		},
	)
	if err != nil {
		return nil, err
	}

	if len(roleDefinitions) == 0 {
		return nil, fmt.Errorf("role definition with Id \"%s\" not found", roleDefinitionId)
	}

	if len(roleDefinitions) > 1 {
		return nil, fmt.Errorf("multiple role definitions with Id \"%s\" found", roleDefinitionId)
	}

	return roleDefinitions[0], nil
}
