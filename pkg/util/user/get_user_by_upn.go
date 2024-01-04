package user

import (
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetUserByUpn(graphServiceClient *msgraphsdkgo.GraphServiceClient, upn string) (models.Userable, error) {
	users, err := GetUsers(
		graphServiceClient,
		func(u models.Userable) bool {
			return *u.GetUserPrincipalName() == upn
		},
	)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with upn \"%s\" not found", upn)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users with upn \"%s\" found", upn)
	}

	return users[0], nil
}
