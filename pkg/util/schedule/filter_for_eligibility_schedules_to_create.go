package schedule

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func FilterForEligibilitySchedulesToCreate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	eligibilitySchedules []*core.Schedule,
	existingRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
) (filtered []*core.Schedule, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	linq.From(eligibilitySchedules).WhereT(func(a *core.Schedule) bool {
		any := linq.From(existingRoleEligibilitySchedules).WhereT(func(s *armauthorization.RoleEligibilitySchedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				scope,
				*s.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				*s.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return a.Scope == *s.Properties.Scope &&
				a.RoleName == *roleDefinition.Properties.RoleName &&
				a.PrincipalName == *principalName
		}).Any()

		return !any
	}).ToSlice(&filtered)

	return
}
