package group_schedule

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func FilterForAssignmentSchedulesToUpdate(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	assignmentSchedules []*core.GroupSchedule,
	existingGroupAssignmentSchedules []models.PrivilegedAccessGroupAssignmentScheduleable,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
) (filtered []*core.GroupSchedule, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	linq.From(assignmentSchedules).WhereT(func(a *core.GroupSchedule) bool {
		idx := linq.From(existingGroupAssignmentSchedules).IndexOfT(func(s models.PrivilegedAccessGroupAssignmentScheduleable) bool {
			accessId := s.GetAccessId()
			var roleName string
			switch *accessId {
			case models.MEMBER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
				roleName = "Member"
			case models.OWNER_PRIVILEGEDACCESSGROUPRELATIONSHIPS:
				roleName = "Owner"
			default:
				panic(fmt.Errorf("accessId with value \"%s\" not supported", accessId))
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				*s.GetPrincipalId(),
			)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, *s.GetGroupId())
			if err != nil {
				panic(err)
			}

			return a.ManagedGroupName == *group.GetDisplayName() &&
				a.RoleName == roleName &&
				a.PrincipalName == *principalName
		})
		if idx == -1 {
			return false
		}

		existingGroupAssignmentSchedule := existingGroupAssignmentSchedules[idx]

		// If start time in config is nil, then we don't want to update the start time in Azure
		// because it will be set to when the schedule was created, which is fine.
		if a.StartDateTime != nil {
			// If there is a start time in config, compare to Azure and flag for update as needed.
			if *existingGroupAssignmentSchedule.GetScheduleInfo().GetStartDateTime() != *a.StartDateTime {
				return true
			}
		}

		// If end date is present in config and Azure, compare and flag for update as needed.
		if existingGroupAssignmentSchedule.GetScheduleInfo().GetExpiration().GetEndDateTime() != nil && a.EndDateTime != nil {
			if *existingGroupAssignmentSchedule.GetScheduleInfo().GetExpiration().GetEndDateTime() != *a.EndDateTime {
				return true
			}
		} else if (existingGroupAssignmentSchedule.GetScheduleInfo().GetExpiration().GetEndDateTime() != nil && a.EndDateTime == nil) || (existingGroupAssignmentSchedule.GetScheduleInfo().GetExpiration().GetEndDateTime() == nil && a.EndDateTime != nil) {
			// If end date is present in config but not Azure, or vice versa, flag for update.
			return true
		}

		return false
	}).ToSlice(&filtered)

	return
}
