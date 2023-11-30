package group

import (
	"fmt"
	"slices"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetGroupById(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache, groupId string) (models.Groupable, error) {
	groups, err := GetGroups(graphServiceClient, cache)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(groups, func(g models.Groupable) bool {
		return *g.GetId() == groupId
	})

	if idx == -1 {
		return nil, fmt.Errorf("group with Id \"%s\" not found", groupId)
	}

	return groups[idx], nil
}
