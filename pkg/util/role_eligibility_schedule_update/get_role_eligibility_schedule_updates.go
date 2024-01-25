package role_eligibility_schedule_update

import (
	"fmt"
	"slices"
	"time"

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

func GetRoleEligibilityScheduleUpdates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupEligibilitySchedules []*core.Schedule,
	existingGroupRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	userEligibilitySchedules []*core.Schedule,
	existingUserRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleUpdate, error) {
	var roleEligibilityScheduleUpdates []*core.RoleEligibilityScheduleUpdate

	groupEligibilitySchedulesToUpdate, err := schedule.FilterForEligibilitySchedulesToUpdate(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibilitySchedules,
		existingGroupRoleEligibilitySchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupEligibilitySchedulesToUpdate {
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

		existingGroupRoleEligibilityScheduleIdx := slices.IndexFunc(existingGroupRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *group.GetId()
		})
		if existingGroupRoleEligibilityScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role eligibility schedule not found")
		}

		existingGroupRoleEligibilitySchedule := existingGroupRoleEligibilitySchedules[existingGroupRoleEligibilityScheduleIdx]

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingGroupRoleEligibilitySchedule.Properties.StartDateTime
		}

		scheduleInfo := schedule_info.GetRoleEligibilityScheduleInfo(startTime, a.EndDateTime)
		roleEligibilityScheduleUpdates = append(roleEligibilityScheduleUpdates, &core.RoleEligibilityScheduleUpdate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              *existingGroupRoleEligibilitySchedule.Properties.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	userEligibilitySchedulesToUpdate, err := schedule.FilterForEligibilitySchedulesToUpdate(
		clientFactory,
		graphServiceClient,
		scope,
		userEligibilitySchedules,
		existingUserRoleEligibilitySchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userEligibilitySchedulesToUpdate {
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

		existingUserRoleEligibilityScheduleIdx := slices.IndexFunc(existingUserRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *user.GetId()
		})
		if existingUserRoleEligibilityScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role eligibility schedule not found")
		}

		existingUserRoleEligibilitySchedule := existingUserRoleEligibilitySchedules[existingUserRoleEligibilityScheduleIdx]

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingUserRoleEligibilitySchedule.Properties.StartDateTime
		}

		scheduleInfo := schedule_info.GetRoleEligibilityScheduleInfo(startTime, a.EndDateTime)
		roleEligibilityScheduleUpdates = append(roleEligibilityScheduleUpdates, &core.RoleEligibilityScheduleUpdate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					Justification:    to.Ptr("Managed by Sheriff"),
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              *existingUserRoleEligibilitySchedule.Properties.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	return roleEligibilityScheduleUpdates, nil
}
