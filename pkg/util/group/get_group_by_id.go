package group

import (
	"context"
	"fmt"
	"strings"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func GetGroupById(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupId string) (models.Groupable, error) {
	var group models.Groupable
	cacheKey := fmt.Sprintf("group_%s", groupId)

	if g, found := cache.Get(cacheKey); found {
		group = g.(models.Groupable)
	} else {
		result, err := graphServiceClient.Groups().ByGroupId(groupId).Get(context.Background(), nil)
		if err != nil {
			if strings.HasPrefix(err.Error(), fmt.Sprintf("Resource '%s' does not exist", groupId)) {
				return nil, fmt.Errorf("group with Id \"%s\" not found", groupId)
			} else {
				return nil, err
			}
		}

		cache.Set(cacheKey, result, gocache.NoExpiration)

		group = result
	}

	return group, nil
}
