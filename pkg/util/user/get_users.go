package user

import (
	"context"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func GetUsers(graphServiceClient *msgraphsdkgo.GraphServiceClient, filter func(models.Userable) bool) ([]models.Userable, error) {
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

	if filter != nil {
		var filteredUsers []models.Userable
		for _, u := range users {
			if filter(u) {
				filteredUsers = append(filteredUsers, u)
			}
		}

		return filteredUsers, nil
	}

	return users, nil
}
