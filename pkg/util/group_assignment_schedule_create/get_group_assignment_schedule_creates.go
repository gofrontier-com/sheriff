package group_assignment_schedule_create

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/group_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/request_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupAssignmentScheduleCreates(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	groupAssignmentSchedules []*core.GroupSchedule,
	existingGroupGroupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
	userAssignmentSchedules []*core.GroupSchedule,
	existingUserGroupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
) ([]*core.GroupAssignmentScheduleCreate, error) {
	var groupAssignmentScheduleCreates []*core.GroupAssignmentScheduleCreate

	groupAssignmentSchedulesToCreate, err := group_schedule.FilterForAssignmentSchedulesToCreate(
		graphServiceClient,
		groupAssignmentSchedules,
		existingGroupGroupAssignmentSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupAssignmentSchedulesToCreate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.ManagedGroupName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		var accessId models.PrivilegedAccessGroupRelationships
		switch a.RoleName {
		case "Member":
			accessId = models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		case "Owner":
			accessId = models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		default:
			panic(fmt.Errorf("role name with value \"%s\" not supported", a.RoleName))
		}

		requestSchedule := request_schedule.GetGroupAssignmentRequestSchedule(a.StartDateTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupAssignmentScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINASSIGN_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(group.GetId())
		scheduleRequest.SetScheduleInfo(requestSchedule)

		groupAssignmentScheduleCreates = append(groupAssignmentScheduleCreates, &core.GroupAssignmentScheduleCreate{
			EndDateTime:                    requestSchedule.GetExpiration().GetEndDateTime(),
			ManagedGroupName:               a.ManagedGroupName,
			PrincipalName:                  *group.GetDisplayName(),
			PrincipalType:                  armauthorization.PrincipalTypeGroup,
			GroupAssignmentScheduleRequest: scheduleRequest,
			RoleName:                       a.RoleName,
			StartDateTime:                  requestSchedule.GetStartDateTime(),
		})
	}

	userAssignmentSchedulesToCreate, err := group_schedule.FilterForAssignmentSchedulesToCreate(
		graphServiceClient,
		userAssignmentSchedules,
		existingUserGroupAssignmentSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userAssignmentSchedulesToCreate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.ManagedGroupName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		var accessId models.PrivilegedAccessGroupRelationships
		switch a.RoleName {
		case "Member":
			accessId = models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		case "Owner":
			accessId = models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS
		default:
			panic(fmt.Errorf("role name with value \"%s\" not supported", a.RoleName))
		}

		requestSchedule := request_schedule.GetGroupAssignmentRequestSchedule(a.StartDateTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupAssignmentScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINASSIGN_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(user.GetId())
		scheduleRequest.SetScheduleInfo(requestSchedule)

		groupAssignmentScheduleCreates = append(groupAssignmentScheduleCreates, &core.GroupAssignmentScheduleCreate{
			EndDateTime:                    requestSchedule.GetExpiration().GetEndDateTime(),
			ManagedGroupName:               a.ManagedGroupName,
			PrincipalName:                  *user.GetUserPrincipalName(),
			PrincipalType:                  armauthorization.PrincipalTypeUser,
			GroupAssignmentScheduleRequest: scheduleRequest,
			RoleName:                       a.RoleName,
			StartDateTime:                  requestSchedule.GetStartDateTime(),
		})
	}

	return groupAssignmentScheduleCreates, nil
}
