package group

import (
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupByName(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupName string) (models.Groupable, error) {
	groups, err := GetGroups(
		graphServiceClient,
		func(g models.Groupable) bool {
			return *g.GetDisplayName() == groupName
		},
	)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("group with name \"%s\" not found", groupName)
	}

	if len(groups) > 1 {
		return nil, fmt.Errorf("multiple groups with name \"%s\" found", groupName)
	}

	return groups[0], nil
}
