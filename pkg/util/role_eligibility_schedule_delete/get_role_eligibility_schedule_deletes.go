package role_eligibility_schedule_delete

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	"github.com/gofrontier-com/sheriff/pkg/util/role_eligibility_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetRoleEligibilityScheduleDeletes(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupSchedules []*core.Schedule,
	existingGroupSchedules []*armauthorization.RoleEligibilitySchedule,
	userSchedules []*core.Schedule,
	existingUserSchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleDelete, error) {
	var roleEligibilityScheduleDeletes []*core.RoleEligibilityScheduleDelete

	groupSchedulesToDelete, err := role_eligibility_schedule.FilterForRoleEligibilitySchedulesToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingGroupSchedules,
		groupSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range groupSchedulesToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			*s.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupById(graphServiceClient, *s.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		if *s.Properties.Status == armauthorization.StatusProvisioned {
			roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, &core.RoleEligibilityScheduleDelete{
				Cancel:        false,
				EndDateTime:   s.Properties.EndDateTime,
				PrincipalName: *group.GetDisplayName(),
				PrincipalType: armauthorization.PrincipalTypeGroup,
				RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
					Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
						Justification:                   to.Ptr("Managed by Sheriff"),
						PrincipalID:                     s.Properties.PrincipalID,
						RequestType:                     to.Ptr(armauthorization.RequestTypeAdminRemove),
						RoleDefinitionID:                s.Properties.RoleDefinitionID,
						TargetRoleEligibilityScheduleID: s.ID,
					},
				},
				RoleEligibilityScheduleRequestName: uuid.New().String(),
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
				StartDateTime:                      s.Properties.StartDateTime,
			})
		} else {
			roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, &core.RoleEligibilityScheduleDelete{
				Cancel:                             true,
				EndDateTime:                        s.Properties.EndDateTime,
				PrincipalName:                      *group.GetDisplayName(),
				PrincipalType:                      armauthorization.PrincipalTypeGroup,
				RoleEligibilityScheduleRequestName: *s.Name,
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
				StartDateTime:                      s.Properties.StartDateTime,
			})
		}
	}

	userSchedulesToDelete, err := role_eligibility_schedule.FilterForRoleEligibilitySchedulesToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingUserSchedules,
		userSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range userSchedulesToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			*s.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserById(graphServiceClient, *s.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		if *s.Properties.Status == armauthorization.StatusProvisioned {
			roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, &core.RoleEligibilityScheduleDelete{
				Cancel:        false,
				EndDateTime:   s.Properties.EndDateTime,
				PrincipalName: *user.GetUserPrincipalName(),
				PrincipalType: armauthorization.PrincipalTypeUser,
				RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
					Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
						Justification:                   to.Ptr("Managed by Sheriff"),
						PrincipalID:                     s.Properties.PrincipalID,
						RequestType:                     to.Ptr(armauthorization.RequestTypeAdminRemove),
						RoleDefinitionID:                s.Properties.RoleDefinitionID,
						TargetRoleEligibilityScheduleID: s.ID,
					},
				},
				RoleEligibilityScheduleRequestName: uuid.New().String(),
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
				StartDateTime:                      s.Properties.StartDateTime,
			})
		} else {
			roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, &core.RoleEligibilityScheduleDelete{
				Cancel:                             true,
				EndDateTime:                        s.Properties.EndDateTime,
				PrincipalName:                      *user.GetUserPrincipalName(),
				PrincipalType:                      armauthorization.PrincipalTypeUser,
				RoleEligibilityScheduleRequestName: *s.Name,
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
				StartDateTime:                      s.Properties.StartDateTime,
			})
		}
	}

	return roleEligibilityScheduleDeletes, nil
}
