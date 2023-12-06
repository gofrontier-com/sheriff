package user

import (
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	gocache "github.com/patrickmn/go-cache"
)

func GetUserUpnById(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache, principalId string) (*string, error) {
	user, err := GetUserById(graphServiceClient, cache, principalId)
	if err != nil {
		return nil, err
	}
	return user.GetUserPrincipalName(), nil
}
