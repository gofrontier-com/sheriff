package user

import (
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetUserUpnById(graphServiceClient *msgraphsdkgo.GraphServiceClient, principalId string) (*string, error) {
	user, err := GetUserById(graphServiceClient, principalId)
	if err != nil {
		return nil, err
	}
	return user.GetUserPrincipalName(), nil
}
