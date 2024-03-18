package directory_object

import (
	"context"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetDirectoryObject(graphServiceClient *msgraphsdkgo.GraphServiceClient, objectId string) (models.DirectoryObjectable, error) {
	return graphServiceClient.DirectoryObjects().ByDirectoryObjectId(objectId).Get(context.Background(), nil)
}
