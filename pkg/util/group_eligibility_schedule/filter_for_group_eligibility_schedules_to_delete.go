package group_eligibility_schedule

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func FilterForEligibilitySchedulesToDelete(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	existingGroupEligibilitySchedules []models.PrivilegedAccessGroupEligibilityScheduleable,
	eligibilitySchedules []*core.Schedule,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
) (filtered []models.PrivilegedAccessGroupEligibilityScheduleable, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	linq.From(existingGroupEligibilitySchedules).WhereT(func(r models.PrivilegedAccessGroupEligibilityScheduleable) bool {
		any := linq.From(eligibilitySchedules).WhereT(func(a *core.Schedule) bool {
			accessId := r.GetAccessId()
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
				*r.GetPrincipalId(),
			)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, *r.GetGroupId())
			if err != nil {
				panic(err)
			}

			return *group.GetDisplayName() == a.Target &&
				roleName == a.RoleName &&
				*principalName == a.PrincipalName
		}).Any()

		return !any
	}).ToSlice(&filtered)

	return
}
