package user

import (
	"fmt"
	"slices"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetUserById(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache, userId string) (models.Userable, error) {
	users, err := GetUsers(graphServiceClient, cache)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(users, func(u models.Userable) bool {
		return *u.GetId() == userId
	})

	if idx == -1 {
		return nil, fmt.Errorf("user with id \"%s\" not found", userId)
	}

	return users[idx], nil
}
