package user

import (
	"context"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetUsers(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache) ([]models.Userable, error) {
	var users = []models.Userable{}
	cacheKey := "users"

	if u, found := cache.Get(cacheKey); found {
		users = u.([]models.Userable)
	} else {

		response, err := graphServiceClient.Users().Get(context.Background(), nil)
		if err != nil {
			return nil, err
		}

		users = response.GetValue()

		cache.Set(cacheKey, users, gocache.NoExpiration)
	}

	return users, nil
}
