package group

import (
	"context"
	"fmt"
	"strings"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetGroupById(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupId string) (models.Groupable, error) {
	var group models.Groupable
	cacheKey := fmt.Sprintf("id::%s", groupId)

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

		group = result

		cacheKeys := []string{
			cacheKey,
			fmt.Sprintf("name::%s", *group.GetDisplayName()),
		}
		for _, cacheKey := range cacheKeys {
			cache.Set(cacheKey, group, gocache.NoExpiration)
		}
	}

	return group, nil
}
