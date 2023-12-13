package role_definition

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func GetRoleDefinitionByName(clientFactory *armauthorization.ClientFactory, scope string, roleDefinitionName string) (*armauthorization.RoleDefinition, error) {
	roleDefinitions, err := GetRoleDefinitions(
		clientFactory,
		scope,
		func(r *armauthorization.RoleDefinition) bool {
			return *r.Properties.RoleName == roleDefinitionName
		},
	)
	if err != nil {
		return nil, err
	}

	if len(roleDefinitions) == 0 {
		return nil, fmt.Errorf("role definition with name \"%s\" not found", roleDefinitionName)
	}

	if len(roleDefinitions) > 1 {
		return nil, fmt.Errorf("multiple role definition with name \"%s\" found", roleDefinitionName)
	}

	return roleDefinitions[0], nil
}
