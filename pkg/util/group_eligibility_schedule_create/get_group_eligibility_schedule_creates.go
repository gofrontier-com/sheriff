package group_eligibility_schedule_create

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/schedule_info"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupEligibilityScheduleCreates(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	groupSchedules []*core.Schedule,
	existingGroupSchedules []models.PrivilegedAccessGroupEligibilityScheduleable,
	userSchedules []*core.Schedule,
	existingUserSchedules []models.PrivilegedAccessGroupEligibilityScheduleable,
) ([]*core.GroupEligibilityScheduleCreate, error) {
	var groupEligibilityScheduleCreates []*core.GroupEligibilityScheduleCreate

	groupSchedulesToCreate, err := schedule.FilterForGroupEligibilitySchedulesToCreate(
		graphServiceClient,
		groupSchedules,
		existingGroupSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupSchedulesToCreate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.Target)
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

		scheduleInfo := schedule_info.GetGroupScheduleInfo(a.StartDateTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupEligibilityScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINASSIGN_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(group.GetId())
		scheduleRequest.SetScheduleInfo(scheduleInfo)

		groupEligibilityScheduleCreates = append(groupEligibilityScheduleCreates, &core.GroupEligibilityScheduleCreate{
			EndDateTime:                     scheduleInfo.GetExpiration().GetEndDateTime(),
			ManagedGroupName:                a.Target,
			PrincipalName:                   *group.GetDisplayName(),
			PrincipalType:                   armauthorization.PrincipalTypeGroup,
			GroupEligibilityScheduleRequest: scheduleRequest,
			RoleName:                        a.RoleName,
			StartDateTime:                   scheduleInfo.GetStartDateTime(),
		})
	}

	userSchedulesToCreate, err := schedule.FilterForGroupEligibilitySchedulesToCreate(
		graphServiceClient,
		userSchedules,
		existingUserSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userSchedulesToCreate {
		managedGroup, err := group.GetGroupByName(graphServiceClient, a.Target)
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

		scheduleInfo := schedule_info.GetGroupScheduleInfo(a.StartDateTime, a.EndDateTime)

		scheduleRequest := models.NewPrivilegedAccessGroupEligibilityScheduleRequest()
		scheduleRequest.SetAccessId(&accessId)
		scheduleRequest.SetAction(to.Ptr(models.ADMINASSIGN_SCHEDULEREQUESTACTIONS))
		scheduleRequest.SetGroupId(managedGroup.GetId())
		scheduleRequest.SetJustification(to.Ptr("Managed by Sheriff"))
		scheduleRequest.SetPrincipalId(user.GetId())
		scheduleRequest.SetScheduleInfo(scheduleInfo)

		groupEligibilityScheduleCreates = append(groupEligibilityScheduleCreates, &core.GroupEligibilityScheduleCreate{
			EndDateTime:                     scheduleInfo.GetExpiration().GetEndDateTime(),
			ManagedGroupName:                a.Target,
			PrincipalName:                   *user.GetUserPrincipalName(),
			PrincipalType:                   armauthorization.PrincipalTypeUser,
			GroupEligibilityScheduleRequest: scheduleRequest,
			RoleName:                        a.RoleName,
			StartDateTime:                   scheduleInfo.GetStartDateTime(),
		})
	}

	return groupEligibilityScheduleCreates, nil
}
