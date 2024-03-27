package group_eligibility_schedule_update

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule_info"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupEligibilityScheduleUpdates(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	groupSchedules []*core.Schedule,
	existingGroupSchedules []models.PrivilegedAccessGroupEligibilityScheduleable,
	userSchedules []*core.Schedule,
	existingUserSchedules []models.PrivilegedAccessGroupEligibilityScheduleable,
) ([]*core.GroupEligibilityScheduleUpdate, error) {
	var groupEligibilityScheduleUpdates []*core.GroupEligibilityScheduleUpdate

	groupSchedulesToUpdate, err := schedule.FilterForGroupEligibilitySchedulesToUpdate(
		graphServiceClient,
		groupSchedules,
		existingGroupSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupSchedulesToUpdate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.Target)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		existingScheduleIdx := linq.From(existingGroupSchedules).IndexOfT(func(s models.PrivilegedAccessGroupEligibilityScheduleable) bool {
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
		if existingScheduleIdx == -1 {
			return nil, fmt.Errorf("existing group eligibility schedule not found")
		}

		existingSchedule := existingGroupSchedules[existingScheduleIdx]

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
			startTime = existingSchedule.GetScheduleInfo().GetStartDateTime()
		}

		scheduleInfo := schedule_info.GetGroupScheduleInfo(startTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupEligibilityScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINUPDATE_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(group.GetId())
		scheduleRequest.SetScheduleInfo(scheduleInfo)

		groupEligibilityScheduleUpdates = append(groupEligibilityScheduleUpdates, &core.GroupEligibilityScheduleUpdate{
			EndDateTime:                     scheduleInfo.GetExpiration().GetEndDateTime(),
			ManagedGroupName:                *managedGroup.GetDisplayName(), // TODO: Not consistent with role approach, but okay?
			PrincipalName:                   *group.GetDisplayName(),
			PrincipalType:                   armauthorization.PrincipalTypeGroup,
			GroupEligibilityScheduleRequest: scheduleRequest,
			RoleName:                        a.RoleName,
			StartDateTime:                   scheduleInfo.GetStartDateTime(),
		})
	}

	userSchedulesToUpdate, err := schedule.FilterForGroupEligibilitySchedulesToUpdate(
		graphServiceClient,
		userSchedules,
		existingUserSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userSchedulesToUpdate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.Target)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		existingScheduleIdx := linq.From(existingUserSchedules).IndexOfT(func(s models.PrivilegedAccessGroupEligibilityScheduleable) bool {
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
		if existingScheduleIdx == -1 {
			return nil, fmt.Errorf("existing group eligibility schedule not found")
		}

		existingSchedule := existingUserSchedules[existingScheduleIdx]

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
			startTime = existingSchedule.GetScheduleInfo().GetStartDateTime()
		}

		scheduleInfo := schedule_info.GetGroupScheduleInfo(startTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupEligibilityScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINUPDATE_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(user.GetId())
		scheduleRequest.SetScheduleInfo(scheduleInfo)

		groupEligibilityScheduleUpdates = append(groupEligibilityScheduleUpdates, &core.GroupEligibilityScheduleUpdate{
			EndDateTime:                     scheduleInfo.GetExpiration().GetEndDateTime(),
			ManagedGroupName:                *managedGroup.GetDisplayName(), // TODO: Not consistent with role approach, but okay?
			PrincipalName:                   *user.GetUserPrincipalName(),
			PrincipalType:                   armauthorization.PrincipalTypeUser,
			GroupEligibilityScheduleRequest: scheduleRequest,
			RoleName:                        a.RoleName,
			StartDateTime:                   scheduleInfo.GetStartDateTime(),
		})
	}

	return groupEligibilityScheduleUpdates, nil
}