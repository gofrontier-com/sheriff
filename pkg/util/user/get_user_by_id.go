package user

import (
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetUserById(graphServiceClient *msgraphsdkgo.GraphServiceClient, userId string) (models.Userable, error) {
	users, err := GetUsers(
		graphServiceClient,
		func(u models.Userable) bool {
			return *u.GetId() == userId
		},
	)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with Id \"%s\" not found", userId)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users with Id \"%s\" found", userId)
	}

	return users[0], nil
}
