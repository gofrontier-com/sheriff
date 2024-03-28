package group_assignment_schedule_delete

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/group_assignment_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupAssignmentScheduleDeletes(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	groupSchedules []*core.Schedule,
	existingGroupSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
	userSchedules []*core.Schedule,
	existingUserSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
) ([]*core.GroupAssignmentScheduleDelete, error) {
	var groupAssignmentScheduleDeletes []*core.GroupAssignmentScheduleDelete

	groupSchedulesToDelete, err := group_assignment_schedule.FilterForAssignmentSchedulesToDelete(
		graphServiceClient,
		existingGroupSchedules,
		groupSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range groupSchedulesToDelete {
		managedGroup, err := group.GetGroupById(graphServiceClient, *s.GetGroupId())
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupById(graphServiceClient, *s.GetPrincipalId())
		if err != nil {
			return nil, err
		}

		var roleName string
		switch *s.GetAccessId() {
		case models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
			roleName = "Member"
		case models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
			roleName = "Owner"
		default:
			panic(fmt.Errorf("accessId with value \"%s\" not supported", *s.GetAccessId()))
		}

		scheduleRequest := models.NewPrivilegedAccessGroupAssignmentScheduleRequest()
		scheduleRequest.SetAccessId(s.GetAccessId())
		scheduleRequest.SetAction(to.Ptr(models.ADMINREMOVE_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(s.GetGroupId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(s.GetPrincipalId())
		scheduleRequest.SetTargetScheduleId(s.GetId())

		if *s.GetStatus() == "Provisioned" {
			groupAssignmentScheduleDeletes = append(groupAssignmentScheduleDeletes, &core.GroupAssignmentScheduleDelete{
				Cancel:                         false,
				EndDateTime:                    s.GetScheduleInfo().GetExpiration().GetEndDateTime(),
				ManagedGroupName:               *managedGroup.GetDisplayName(),
				PrincipalName:                  *group.GetDisplayName(),
				PrincipalType:                  armauthorization.PrincipalTypeGroup,
				GroupAssignmentScheduleRequest: scheduleRequest,
				RoleName:                       roleName,
				StartDateTime:                  s.GetScheduleInfo().GetStartDateTime(),
			})
		} else {
			groupAssignmentScheduleDeletes = append(groupAssignmentScheduleDeletes, &core.GroupAssignmentScheduleDelete{
				Cancel:           true,
				EndDateTime:      s.GetScheduleInfo().GetExpiration().GetEndDateTime(),
				ManagedGroupName: *managedGroup.GetDisplayName(),
				PrincipalName:    *group.GetDisplayName(),
				PrincipalType:    armauthorization.PrincipalTypeGroup,
				RoleName:         roleName,
				StartDateTime:    s.GetScheduleInfo().GetStartDateTime(),
			})
		}
	}

	userSchedulesToDelete, err := group_assignment_schedule.FilterForAssignmentSchedulesToDelete(
		graphServiceClient,
		existingUserSchedules,
		userSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range userSchedulesToDelete {
		managedGroup, err := group.GetGroupById(graphServiceClient, *s.GetGroupId())
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserById(graphServiceClient, *s.GetPrincipalId())
		if err != nil {
			return nil, err
		}

		var roleName string
		switch *s.GetAccessId() {
		case models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
			roleName = "Member"
		case models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
			roleName = "Owner"
		default:
			panic(fmt.Errorf("accessId with value \"%s\" not supported", *s.GetAccessId()))
		}

		scheduleRequest := models.NewPrivilegedAccessGroupAssignmentScheduleRequest()
		scheduleRequest.SetAccessId(s.GetAccessId())
		scheduleRequest.SetAction(to.Ptr(models.ADMINREMOVE_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(s.GetGroupId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(s.GetPrincipalId())
		scheduleRequest.SetTargetScheduleId(s.GetId())

		if *s.GetStatus() == "Provisioned" {
			groupAssignmentScheduleDeletes = append(groupAssignmentScheduleDeletes, &core.GroupAssignmentScheduleDelete{
				Cancel:                         false,
				EndDateTime:                    s.GetScheduleInfo().GetExpiration().GetEndDateTime(),
				ManagedGroupName:               *managedGroup.GetDisplayName(),
				PrincipalName:                  *user.GetUserPrincipalName(),
				PrincipalType:                  armauthorization.PrincipalTypeUser,
				GroupAssignmentScheduleRequest: scheduleRequest,
				RoleName:                       roleName,
				StartDateTime:                  s.GetScheduleInfo().GetStartDateTime(),
			})
		} else {
			groupAssignmentScheduleDeletes = append(groupAssignmentScheduleDeletes, &core.GroupAssignmentScheduleDelete{
				Cancel:           true,
				EndDateTime:      s.GetScheduleInfo().GetExpiration().GetEndDateTime(),
				ManagedGroupName: *managedGroup.GetDisplayName(),
				PrincipalName:    *user.GetUserPrincipalName(),
				PrincipalType:    armauthorization.PrincipalTypeUser,
				RoleName:         roleName,
				StartDateTime:    s.GetScheduleInfo().GetStartDateTime(),
			})
		}
	}

	return groupAssignmentScheduleDeletes, nil
}
