package group

import (
	"context"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetGroups(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache) ([]models.Groupable, error) {
	var groups = []models.Groupable{}
	cacheKey := "groups"

	if g, found := cache.Get(cacheKey); found {
		groups = g.([]models.Groupable)
	} else {

		response, err := graphServiceClient.Groups().Get(context.Background(), nil)
		if err != nil {
			return nil, err
		}

		groups = response.GetValue()

		cache.Set(cacheKey, groups, gocache.NoExpiration)
	}

	return groups, nil
}
