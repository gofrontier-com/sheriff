package role_definition

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	gocache "github.com/patrickmn/go-cache"
)

func GetRoleDefinitionById(clientFactory *armauthorization.ClientFactory, roleDefinitionId string) (*armauthorization.RoleDefinition, error) {
	var roleDefinition *armauthorization.RoleDefinition
	cacheKey := fmt.Sprintf("id::%s", roleDefinitionId)

	if d, found := cache.Get(cacheKey); found {
		roleDefinition = d.(*armauthorization.RoleDefinition)
	} else {
		roleDefinitionsClient := clientFactory.NewRoleDefinitionsClient()

		options := &armauthorization.RoleDefinitionsClientGetByIDOptions{}
		response, err := roleDefinitionsClient.GetByID(context.Background(), roleDefinitionId, options)
		if err != nil {
			if _, ok := err.(*azcore.ResponseError); ok {
				responseError := err.(*azcore.ResponseError)
				if responseError.ErrorCode == "RoleDefinitionDoesNotExist" {
					return nil, fmt.Errorf("role definition with Id \"%s\" not found", roleDefinitionId)
				}
			} else {
				return nil, err
			}
		}

		roleDefinition = &response.RoleDefinition

		scope := strings.Split(*roleDefinition.ID, "/providers/Microsoft.Authorization/roleDefinitions/")[0]

		cacheKeys := []string{
			cacheKey,
			fmt.Sprintf("scoped-name::%s:%s", scope, *roleDefinition.Properties.RoleName),
		}
		for _, cacheKey := range cacheKeys {
			cache.Set(cacheKey, roleDefinition, gocache.NoExpiration)
		}
	}

	return roleDefinition, nil
}
