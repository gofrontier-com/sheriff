package role_assignment_schedule_update

import (
	"fmt"
	"slices"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule_info"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetRoleAssignmentScheduleUpdates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupAssignmentSchedules []*core.Schedule,
	existingGroupRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
	userAssignmentSchedules []*core.Schedule,
	existingUserRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
) ([]*core.RoleAssignmentScheduleUpdate, error) {
	var roleAssignmentScheduleUpdates []*core.RoleAssignmentScheduleUpdate

	groupAssignmentSchedulesToUpdate, err := schedule.FilterForAssignmentSchedulesToUpdate(
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

	for _, a := range groupAssignmentSchedulesToUpdate {
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

		existingGroupRoleAssignmentScheduleIdx := slices.IndexFunc(existingGroupRoleAssignmentSchedules, func(s *armauthorization.RoleAssignmentSchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *group.GetId()
		})
		existingGroupRoleAssignmentScheduleIdx2 := linq.From(existingGroupRoleAssignmentSchedules).IndexOfT(func(s *armauthorization.RoleAssignmentSchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *group.GetId()
		})
		if existingGroupRoleAssignmentScheduleIdx != existingGroupRoleAssignmentScheduleIdx2 {
			panic("index mismatch")
		}
		if existingGroupRoleAssignmentScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role assignment schedule not found")
		}

		existingGroupRoleAssignmentSchedule := existingGroupRoleAssignmentSchedules[existingGroupRoleAssignmentScheduleIdx]

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingGroupRoleAssignmentSchedule.Properties.StartDateTime
		}

		scheduleInfo := schedule_info.GetRoleAssignmentScheduleInfo(startTime, a.EndDateTime)
		roleAssignmentScheduleUpdates = append(roleAssignmentScheduleUpdates, &core.RoleAssignmentScheduleUpdate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleAssignmentScheduleRequest: &armauthorization.RoleAssignmentScheduleRequest{
				Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleAssignmentScheduleRequestName: uuid.New().String(),
			RoleName:                          *roleDefinition.Properties.RoleName,
			Scope:                             *existingGroupRoleAssignmentSchedule.Properties.Scope,
			StartDateTime:                     scheduleInfo.StartDateTime,
		})
	}

	userAssignmentSchedulesToUpdate, err := schedule.FilterForAssignmentSchedulesToUpdate(
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

	for _, a := range userAssignmentSchedulesToUpdate {
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

		existingUserRoleAssignmentScheduleIdx := slices.IndexFunc(existingUserRoleAssignmentSchedules, func(s *armauthorization.RoleAssignmentSchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *user.GetId()
		})
		existingUserRoleAssignmentScheduleIdx2 := linq.From(existingUserRoleAssignmentSchedules).IndexOfT(func(s *armauthorization.RoleAssignmentSchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *user.GetId()
		})
		if existingUserRoleAssignmentScheduleIdx != existingUserRoleAssignmentScheduleIdx2 {
			panic("index mismatch")
		}
		if existingUserRoleAssignmentScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role assignment schedule not found")
		}

		existingUserRoleAssignmentSchedule := existingUserRoleAssignmentSchedules[existingUserRoleAssignmentScheduleIdx]

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingUserRoleAssignmentSchedule.Properties.StartDateTime
		}

		scheduleInfo := schedule_info.GetRoleAssignmentScheduleInfo(startTime, a.EndDateTime)
		roleAssignmentScheduleUpdates = append(roleAssignmentScheduleUpdates, &core.RoleAssignmentScheduleUpdate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleAssignmentScheduleRequest: &armauthorization.RoleAssignmentScheduleRequest{
				Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleAssignmentScheduleRequestName: uuid.New().String(),
			RoleName:                          *roleDefinition.Properties.RoleName,
			Scope:                             *existingUserRoleAssignmentSchedule.Properties.Scope,
			StartDateTime:                     scheduleInfo.StartDateTime,
		})
	}

	return roleAssignmentScheduleUpdates, nil
}
