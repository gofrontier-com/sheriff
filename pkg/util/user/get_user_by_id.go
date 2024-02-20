package user

import (
	"context"
	"fmt"
	"strings"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

func GetUserById(graphServiceClient *msgraphsdkgo.GraphServiceClient, userId string) (models.Userable, error) {
	var user models.Userable
	cacheKey := fmt.Sprintf("id::%s", userId)

	if u, found := cache.Get(cacheKey); found {
		user = u.(models.Userable)
	} else {
		result, err := graphServiceClient.Users().ByUserId(userId).Get(context.Background(), nil)
		if err != nil {
			if strings.HasPrefix(err.Error(), fmt.Sprintf("Resource '%s' does not exist", userId)) {
				return nil, fmt.Errorf("user with Id \"%s\" not found", userId)
			} else {
				return nil, err
			}
		}

		user = result

		cacheKeys := []string{
			cacheKey,
			fmt.Sprintf("upn::%s", *user.GetUserPrincipalName()),
		}
		for _, cacheKey := range cacheKeys {
			cache.Set(cacheKey, user, gocache.NoExpiration)
		}
	}

	return user, nil
}
