package schedule

import (
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func FilterForEligibilitySchedulesToUpdate(
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
		idx := slices.IndexFunc(existingRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
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
		})
		idx2 := linq.From(existingRoleEligibilitySchedules).IndexOfT(func(s *armauthorization.RoleEligibilitySchedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
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
		})
		if idx != idx2 {
			panic("index mismatch")
		}
		if idx == -1 {
			return false
		}

		existingRoleEligibilitySchedule := existingRoleEligibilitySchedules[idx]

		// If start time in config is nil, then we don't want to update the start time in Azure
		// because it will be set to when the schedule was created, which is fine.
		if a.StartDateTime != nil {
			// If there is a start time in config, compare to Azure and flag for update as needed.
			if *existingRoleEligibilitySchedule.Properties.StartDateTime != *a.StartDateTime {
				return true
			}
		}

		// If end date is present in config and Azure, compare and flag for update as needed.
		if existingRoleEligibilitySchedule.Properties.EndDateTime != nil && a.EndDateTime != nil {
			if *existingRoleEligibilitySchedule.Properties.EndDateTime != *a.EndDateTime {
				return true
			}
		} else if (existingRoleEligibilitySchedule.Properties.EndDateTime != nil && a.EndDateTime == nil) || (existingRoleEligibilitySchedule.Properties.EndDateTime == nil && a.EndDateTime != nil) {
			// If end date is present in config but not Azure, or vice versa, flag for update.
			return true
		}

		return false
	}).ToSlice(&filtered)

	return
}
