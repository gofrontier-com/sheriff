package group_schedule

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func FilterForAssignmentSchedulesToCreate(
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
		any := linq.From(existingGroupAssignmentSchedules).WhereT(func(s models.PrivilegedAccessGroupAssignmentScheduleable) bool {
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
		}).Any()

		return !any
	}).ToSlice(&filtered)

	return
}
