package role_eligibility_schedule_create

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule_info"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetRoleEligibilityScheduleCreates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupSchedules []*core.Schedule,
	existingGroupSchedules []*armauthorization.RoleEligibilitySchedule,
	userSchedules []*core.Schedule,
	existingUserSchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleCreate, error) {
	var roleEligibilityScheduleCreates []*core.RoleEligibilityScheduleCreate

	groupSchedulesToCreate, err := schedule.FilterForRoleEligibilitySchedulesToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		groupSchedules,
		existingGroupSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupSchedulesToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(
			clientFactory,
			scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		scheduleInfo := schedule_info.GetRoleEligibilityScheduleInfo(a.StartDateTime, a.EndDateTime)
		roleEligibilityScheduleCreates = append(roleEligibilityScheduleCreates, &core.RoleEligibilityScheduleCreate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              a.Target,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	userSchedulesToCreate, err := schedule.FilterForRoleEligibilitySchedulesToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		userSchedules,
		existingUserSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userSchedulesToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(
			clientFactory,
			scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		scheduleInfo := schedule_info.GetRoleEligibilityScheduleInfo(a.StartDateTime, a.EndDateTime)
		roleEligibilityScheduleCreates = append(roleEligibilityScheduleCreates, &core.RoleEligibilityScheduleCreate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeUser,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              a.Target,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	return roleEligibilityScheduleCreates, nil
}
