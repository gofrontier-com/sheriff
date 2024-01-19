package group

import (
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetGroupDisplayNameById(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupId string) (*string, error) {
	group, err := GetGroupById(graphServiceClient, groupId)
	if err != nil {
		return nil, err
	}
	return group.GetDisplayName(), nil
}
