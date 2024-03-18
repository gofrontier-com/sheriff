package group_assignment_schedule_update

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/group_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/request_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupAssignmentScheduleUpdates(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	groupAssignmentSchedules []*core.GroupSchedule,
	existingGroupGroupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
	userAssignmentSchedules []*core.GroupSchedule,
	existingUserGroupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
) ([]*core.GroupAssignmentScheduleUpdate, error) {
	var groupAssignmentScheduleUpdates []*core.GroupAssignmentScheduleUpdate

	groupAssignmentSchedulesToUpdate, err := group_schedule.FilterForAssignmentSchedulesToUpdate(
		graphServiceClient,
		groupAssignmentSchedules,
		existingGroupGroupAssignmentSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupAssignmentSchedulesToUpdate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.ManagedGroupName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		existingGroupGroupAssignmentScheduleIdx := linq.From(existingGroupGroupAssignmentSchedules).IndexOfT(func(s models.PrivilegedAccessGroupAssignmentScheduleable) bool {
			var roleName string
			switch *s.GetAccessId() {
			case models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
				roleName = "Member"
			case models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
				roleName = "Owner"
			default:
				panic(fmt.Errorf("accessId with value \"%s\" not supported", *s.GetAccessId()))
			}
			return *s.GetGroupId() == *managedGroup.GetId() &&
				roleName == a.RoleName &&
				*s.GetPrincipalId() == *group.GetId()
		})
		if existingGroupGroupAssignmentScheduleIdx == -1 {
			return nil, fmt.Errorf("existing group assignment schedule not found")
		}

		existingGroupGroupAssignmentSchedule := existingGroupGroupAssignmentSchedules[existingGroupGroupAssignmentScheduleIdx]

		var accessId models.PrivilegedAccessGroupRelationships
		switch a.RoleName {
		case "Member":
			accessId = models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		case "Owner":
			accessId = models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		default:
			panic(fmt.Errorf("role name with value \"%s\" not supported", a.RoleName))
		}

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingGroupGroupAssignmentSchedule.GetScheduleInfo().GetStartDateTime()
		}

		requestSchedule := request_schedule.GetGroupAssignmentRequestSchedule(startTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupAssignmentScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINUPDATE_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(group.GetId())
		scheduleRequest.SetScheduleInfo(requestSchedule)

		groupAssignmentScheduleUpdates = append(groupAssignmentScheduleUpdates, &core.GroupAssignmentScheduleUpdate{
			EndDateTime:                    requestSchedule.GetExpiration().GetEndDateTime(),
			ManagedGroupName:               *managedGroup.GetDisplayName(), // TODO: Not consistent with role approach, but okay?
			PrincipalName:                  *group.GetDisplayName(),
			PrincipalType:                  armauthorization.PrincipalTypeGroup,
			GroupAssignmentScheduleRequest: scheduleRequest,
			RoleName:                       a.RoleName,
			StartDateTime:                  requestSchedule.GetStartDateTime(),
		})
	}

	userAssignmentSchedulesToUpdate, err := group_schedule.FilterForAssignmentSchedulesToUpdate(
		graphServiceClient,
		userAssignmentSchedules,
		existingUserGroupAssignmentSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userAssignmentSchedulesToUpdate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.ManagedGroupName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		existingUserGroupAssignmentScheduleIdx := linq.From(existingUserGroupAssignmentSchedules).IndexOfT(func(s models.PrivilegedAccessGroupAssignmentScheduleable) bool {
			var roleName string
			switch *s.GetAccessId() {
			case models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
				roleName = "Member"
			case models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
				roleName = "Owner"
			default:
				panic(fmt.Errorf("accessId with value \"%s\" not supported", *s.GetAccessId()))
			}
			return *s.GetGroupId() == *managedGroup.GetId() &&
				roleName == a.RoleName &&
				*s.GetPrincipalId() == *user.GetId()
		})
		if existingUserGroupAssignmentScheduleIdx == -1 {
			return nil, fmt.Errorf("existing group assignment schedule not found")
		}

		existingUserGroupAssignmentSchedule := existingUserGroupAssignmentSchedules[existingUserGroupAssignmentScheduleIdx]

		var accessId models.PrivilegedAccessGroupRelationships
		switch a.RoleName {
		case "Member":
			accessId = models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		case "Owner":
			accessId = models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		default:
			panic(fmt.Errorf("role name with value \"%s\" not supported", a.RoleName))
		}

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingUserGroupAssignmentSchedule.GetScheduleInfo().GetStartDateTime()
		}

		requestSchedule := request_schedule.GetGroupAssignmentRequestSchedule(startTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupAssignmentScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINUPDATE_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(user.GetId())
		scheduleRequest.SetScheduleInfo(requestSchedule)

		groupAssignmentScheduleUpdates = append(groupAssignmentScheduleUpdates, &core.GroupAssignmentScheduleUpdate{
			EndDateTime:                    requestSchedule.GetExpiration().GetEndDateTime(),
			ManagedGroupName:               *managedGroup.GetDisplayName(), // TODO: Not consistent with role approach, but okay?
			PrincipalName:                  *user.GetUserPrincipalName(),
			PrincipalType:                  armauthorization.PrincipalTypeUser,
			GroupAssignmentScheduleRequest: scheduleRequest,
			RoleName:                       a.RoleName,
			StartDateTime:                  requestSchedule.GetStartDateTime(),
		})
	}

	return groupAssignmentScheduleUpdates, nil
}
