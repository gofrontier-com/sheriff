package role_assignment_schedule

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func FilterForRoleAssignmentSchedulesToDelete(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	existingRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
	assignmentSchedules []*core.Schedule,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
) (filtered []*armauthorization.RoleAssignmentSchedule, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	linq.From(existingRoleAssignmentSchedules).WhereT(func(r *armauthorization.RoleAssignmentSchedule) bool {
		any := linq.From(assignmentSchedules).WhereT(func(a *core.Schedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				*r.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return *r.Properties.Scope == a.Scope &&
				*roleDefinition.Properties.RoleName == a.RoleName &&
				*principalName == a.PrincipalName
		}).Any()

		return !any
	}).ToSlice(&filtered)

	return
}
