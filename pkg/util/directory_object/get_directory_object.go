package directory_object

import (
	"context"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

// TODO: Cache

func GetDirectoryObject(graphServiceClient *msgraphsdkgo.GraphServiceClient, objectId string) (models.DirectoryObjectable, error) {
	return graphServiceClient.DirectoryObjects().ByDirectoryObjectId(objectId).Get(context.Background(), nil)
}
