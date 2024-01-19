package group

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

func GetGroups(graphServiceClient *msgraphsdkgo.GraphServiceClient, filter func(models.Groupable) bool) ([]models.Groupable, error) {
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

	if filter != nil {
		var filteredGroups []models.Groupable
		for _, g := range groups {
			if filter(g) {
				filteredGroups = append(filteredGroups, g)
			}
		}

		return filteredGroups, nil
	}

	return groups, nil
}
