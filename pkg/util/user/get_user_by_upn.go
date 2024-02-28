package user

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	gocache "github.com/patrickmn/go-cache"
)

func GetUserByUpn(graphServiceClient *msgraphsdkgo.GraphServiceClient, upn string) (models.Userable, error) {
	var user models.Userable
	cacheKey := fmt.Sprintf("upn::%s", upn)

	if u, found := cache.Get(cacheKey); found {
		user = u.(models.Userable)
	} else {
		filterValue := fmt.Sprintf("userPrincipalName eq '%s'", upn)
		query := users.UsersRequestBuilderGetQueryParameters{
			Filter: &filterValue,
		}
		options := users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &query,
		}
		result, err := graphServiceClient.Users().Get(context.Background(), &options)
		if err != nil {
			return nil, err
		}

		users := result.GetValue()

		if len(users) == 0 {
			return nil, fmt.Errorf("user with upn \"%s\" not found", upn)
		}

		if len(users) > 1 {
			return nil, fmt.Errorf("multiple users with upn \"%s\" found", upn)
		}

		user = users[0]

		cacheKeys := []string{
			cacheKey,
			fmt.Sprintf("id::%s", *user.GetId()),
		}
		for _, cacheKey := range cacheKeys {
			cache.Set(cacheKey, user, gocache.NoExpiration)
		}
	}

	return user, nil
}
