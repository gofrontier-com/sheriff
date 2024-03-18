package group_assignment_schedule

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/identitygovernance"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupAssignmentSchedules(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupIds []string, filter func(*msgraphsdkgo.GraphServiceClient, models.PrivilegedAccessGroupAssignmentScheduleable) bool) ([]models.PrivilegedAccessGroupAssignmentScheduleable, error) {
	var groupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable

	for _, i := range groupIds {
		requestConfiguration := &identitygovernance.PrivilegedAccessGroupAssignmentSchedulesRequestBuilderGetRequestConfiguration{
			QueryParameters: &identitygovernance.PrivilegedAccessGroupAssignmentSchedulesRequestBuilderGetQueryParameters{
				Filter: to.Ptr(fmt.Sprintf("groupId eq '%s'", i)),
			},
		}
		result, err := graphServiceClient.IdentityGovernance().PrivilegedAccess().Group().AssignmentSchedules().Get(context.Background(), requestConfiguration)
		if err != nil {
			return nil, err
		}

		groupAssignmentSchedules = append(groupAssignmentSchedules, result.GetValue()...)
	}

	var filteredGroupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable
	for _, s := range groupAssignmentSchedules {
		if filter(graphServiceClient, s) {
			filteredGroupAssignmentSchedules = append(filteredGroupAssignmentSchedules, s)
		}
	}

	return filteredGroupAssignmentSchedules, nil
}
