package user

import (
	"fmt"
	"slices"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetUserByUpn(graphServiceClient *msgraphsdkgo.GraphServiceClient, cache gocache.Cache, upn string) (models.Userable, error) {
	users, err := GetUsers(graphServiceClient, cache)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(users, func(u models.Userable) bool {
		return *u.GetUserPrincipalName() == upn
	})

	if idx == -1 {
		return nil, fmt.Errorf("user with upn \"%s\" not found", upn)
	}

	return users[idx], nil
}
