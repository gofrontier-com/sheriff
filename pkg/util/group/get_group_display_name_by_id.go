package group

import (
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	gocache "github.com/patrickmn/go-cache"
)

func GetGroupDisplayNameById(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache, groupId string) (*string, error) {
	group, err := GetGroupById(graphServiceClient, cache, groupId)
	if err != nil {
		return nil, err
	}
	return group.GetDisplayName(), nil
}
