package group_eligibility_schedule

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/identitygovernance"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupEligibilitySchedules(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupIds []string, filter func(*msgraphsdkgo.GraphServiceClient, *models.PrivilegedAccessGroupEligibilityScheduleable) bool) ([]*models.PrivilegedAccessGroupEligibilityScheduleable, error) {
	var groupEligibilitySchedules []*models.PrivilegedAccessGroupEligibilityScheduleable

	for _, i := range groupIds {
		requestConfiguration := &identitygovernance.PrivilegedAccessGroupEligibilitySchedulesRequestBuilderGetRequestConfiguration{
			QueryParameters: &identitygovernance.PrivilegedAccessGroupEligibilitySchedulesRequestBuilderGetQueryParameters{
				Filter: to.Ptr(fmt.Sprintf("groupId eq '%s'", i)),
			},
		}
		result, err := graphServiceClient.IdentityGovernance().PrivilegedAccess().Group().EligibilitySchedules().Get(context.Background(), requestConfiguration)
		if err != nil {
			return nil, err
		}

		for _, r := range result.GetValue() {
			groupEligibilitySchedules = append(groupEligibilitySchedules, &r)
		}
	}

	var filteredGroupEligibilitySchedules []*models.PrivilegedAccessGroupEligibilityScheduleable
	for _, s := range groupEligibilitySchedules {
		if filter(graphServiceClient, s) {
			filteredGroupEligibilitySchedules = append(filteredGroupEligibilitySchedules, s)
		}
	}

	return filteredGroupEligibilitySchedules, nil
}
