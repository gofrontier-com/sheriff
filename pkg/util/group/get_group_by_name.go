package group

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetGroupByName(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupName string) (models.Groupable, error) {
	var group models.Groupable
	cacheKey := fmt.Sprintf("group_%s", groupName)

	if g, found := cache.Get(cacheKey); found {
		group = g.(models.Groupable)
	} else {
		filterValue := fmt.Sprintf("displayName eq '%s'", groupName)
		query := groups.GroupsRequestBuilderGetQueryParameters{
			Filter: &filterValue,
		}
		options := groups.GroupsRequestBuilderGetRequestConfiguration{
			QueryParameters: &query,
		}
		result, err := graphServiceClient.Groups().Get(context.Background(), &options)
		if err != nil {
			return nil, err
		}

		groups := result.GetValue()

		if len(groups) == 0 {
			return nil, fmt.Errorf("group with display name \"%s\" not found", groupName)
		}

		if len(groups) > 1 {
			return nil, fmt.Errorf("multiple groups with display name \"%s\" found", groupName)
		}

		cache.Set(cacheKey, groups[0], gocache.NoExpiration)

		group = groups[0]
	}

	return group, nil
}
