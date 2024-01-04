package group

import (
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupById(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupId string) (models.Groupable, error) {
	groups, err := GetGroups(
		graphServiceClient,
		func(g models.Groupable) bool {
			return *g.GetId() == groupId
		},
	)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("group with Id \"%s\" not found", groupId)
	}

	if len(groups) > 1 {
		return nil, fmt.Errorf("multiple groups with Id \"%s\" found", groupId)
	}

	return groups[0], nil
}
