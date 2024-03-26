package role_eligibility_schedule_update

import (
	"fmt"
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

func GetRoleEligibilityScheduleUpdates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupSchedules []*core.Schedule,
	existingGroupSchedules []*armauthorization.RoleEligibilitySchedule,
	userSchedules []*core.Schedule,
	existingUserSchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleUpdate, error) {
	var roleEligibilityScheduleUpdates []*core.RoleEligibilityScheduleUpdate

	groupSchedulesToUpdate, err := schedule.FilterForRoleEligibilitySchedulesToUpdate(
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

	for _, a := range groupSchedulesToUpdate {
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

		existingScheduleIdx := linq.From(existingGroupSchedules).IndexOfT(func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.Scope == a.Target &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *group.GetId()
		})
		if existingScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role eligibility schedule not found")
		}

		existingSchedule := existingGroupSchedules[existingScheduleIdx]

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingSchedule.Properties.StartDateTime
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
			Scope:                              *existingSchedule.Properties.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	userSchedulesToUpdate, err := schedule.FilterForRoleEligibilitySchedulesToUpdate(
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

	for _, a := range userSchedulesToUpdate {
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

		existingScheduleIdx := linq.From(existingUserSchedules).IndexOfT(func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.Scope == a.Target &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *user.GetId()
		})
		if existingScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role eligibility schedule not found")
		}

		existingSchedule := existingUserSchedules[existingScheduleIdx]

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingSchedule.Properties.StartDateTime
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
			Scope:                              *existingSchedule.Properties.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	return roleEligibilityScheduleUpdates, nil
}
