package schedule

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func FilterForGroupEligibilitySchedulesToCreate(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	schedules []*core.Schedule,
	existingSchedules []models.PrivilegedAccessGroupEligibilityScheduleable,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
) (filtered []*core.Schedule, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	linq.From(schedules).WhereT(func(a *core.Schedule) bool {
		any := linq.From(existingSchedules).WhereT(func(s models.PrivilegedAccessGroupEligibilityScheduleable) bool {
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

			return a.Target == *group.GetDisplayName() &&
				a.RoleName == roleName &&
				a.PrincipalName == *principalName
		}).Any()

		return !any
	}).ToSlice(&filtered)

	return
}
