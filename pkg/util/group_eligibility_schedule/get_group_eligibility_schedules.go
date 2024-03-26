package group_eligibility_schedule

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/identitygovernance"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	gocache "github.com/patrickmn/go-cache"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

// TODO: Cache.

func GetGroupEligibilitySchedules(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupIds []string, filter func(*msgraphsdkgo.GraphServiceClient, models.PrivilegedAccessGroupEligibilityScheduleable) bool) ([]models.PrivilegedAccessGroupEligibilityScheduleable, error) {
	var groupEligibilitySchedules []models.PrivilegedAccessGroupEligibilityScheduleable

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

		groupEligibilitySchedules = append(groupEligibilitySchedules, result.GetValue()...)
	}

	var filteredGroupEligibilitySchedules []models.PrivilegedAccessGroupEligibilityScheduleable
	for _, s := range groupEligibilitySchedules {
		if filter(graphServiceClient, s) {
			filteredGroupEligibilitySchedules = append(filteredGroupEligibilitySchedules, s)
		}
	}

	return filteredGroupEligibilitySchedules, nil
}
