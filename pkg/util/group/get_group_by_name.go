package group

import (
	"fmt"
	"slices"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetGroupByName(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache, groupName string) (models.Groupable, error) {
	groups, err := GetGroups(graphServiceClient, cache)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(groups, func(g models.Groupable) bool {
		return *g.GetDisplayName() == groupName
	})

	if idx == -1 {
		return nil, fmt.Errorf("group with name \"%s\" not found", groupName)
	}

	return groups[idx], nil
}
