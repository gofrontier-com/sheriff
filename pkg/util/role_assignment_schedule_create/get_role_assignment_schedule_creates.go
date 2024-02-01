package role_assignment_schedule_create

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

func GetRoleAssignmentScheduleCreates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupAssignmentSchedules []*core.Schedule,
	existingGroupRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
	userAssignmentSchedules []*core.Schedule,
	existingUserRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
) ([]*core.RoleAssignmentScheduleCreate, error) {
	var roleAssignmentScheduleCreates []*core.RoleAssignmentScheduleCreate

	groupAssignmentSchedulesToCreate, err := schedule.FilterForAssignmentSchedulesToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		groupAssignmentSchedules,
		existingGroupRoleAssignmentSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupAssignmentSchedulesToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		scheduleInfo := schedule_info.GetRoleAssignmentScheduleInfo(a.StartDateTime, a.EndDateTime)
		roleAssignmentScheduleCreates = append(roleAssignmentScheduleCreates, &core.RoleAssignmentScheduleCreate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleAssignmentScheduleRequest: &armauthorization.RoleAssignmentScheduleRequest{
				Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleAssignmentScheduleRequestName: uuid.New().String(),
			RoleName:                          *roleDefinition.Properties.RoleName,
			Scope:                             a.Scope,
			StartDateTime:                     scheduleInfo.StartDateTime,
		})
	}

	userAssignmentSchedulesToCreate, err := schedule.FilterForAssignmentSchedulesToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		userAssignmentSchedules,
		existingUserRoleAssignmentSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userAssignmentSchedulesToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		scheduleInfo := schedule_info.GetRoleAssignmentScheduleInfo(a.StartDateTime, a.EndDateTime)
		roleAssignmentScheduleCreates = append(roleAssignmentScheduleCreates, &core.RoleAssignmentScheduleCreate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeUser,
			RoleAssignmentScheduleRequest: &armauthorization.RoleAssignmentScheduleRequest{
				Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleAssignmentScheduleRequestName: uuid.New().String(),
			RoleName:                          *roleDefinition.Properties.RoleName,
			Scope:                             a.Scope,
			StartDateTime:                     scheduleInfo.StartDateTime,
		})
	}

	return roleAssignmentScheduleCreates, nil
}
