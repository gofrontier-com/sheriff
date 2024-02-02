package role_assignment_schedule_delete

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/role_assignment_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetRoleAssignmentScheduleDeletes(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupAssignmentSchedules []*core.Schedule,
	existingGroupRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
	userAssignmentSchedules []*core.Schedule,
	existingUserRoleAssignmentSchedules []*armauthorization.RoleAssignmentSchedule,
) ([]*core.RoleAssignmentScheduleDelete, error) {
	var roleAssignmentScheduleDeletes []*core.RoleAssignmentScheduleDelete

	groupAssignmentSchedulesToDelete, err := role_assignment_schedule.FilterForRoleAssignmentSchedulesToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingGroupRoleAssignmentSchedules,
		groupAssignmentSchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range groupAssignmentSchedulesToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			scope,
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
			roleAssignmentScheduleDeletes = append(roleAssignmentScheduleDeletes, &core.RoleAssignmentScheduleDelete{
				Cancel:        false,
				EndDateTime:   s.Properties.EndDateTime,
				PrincipalName: *group.GetDisplayName(),
				PrincipalType: armauthorization.PrincipalTypeGroup,
				RoleAssignmentScheduleRequest: &armauthorization.RoleAssignmentScheduleRequest{
					Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
						Justification:                  to.Ptr("Managed by Sheriff"),
						PrincipalID:                    s.Properties.PrincipalID,
						RequestType:                    to.Ptr(armauthorization.RequestTypeAdminRemove),
						RoleDefinitionID:               s.Properties.RoleDefinitionID,
						TargetRoleAssignmentScheduleID: s.ID,
					},
				},
				RoleAssignmentScheduleRequestName: uuid.New().String(),
				RoleName:                          *roleDefinition.Properties.RoleName,
				Scope:                             *s.Properties.Scope,
				StartDateTime:                     s.Properties.StartDateTime,
			})
		} else {
			roleAssignmentScheduleDeletes = append(roleAssignmentScheduleDeletes, &core.RoleAssignmentScheduleDelete{
				Cancel:                            true,
				EndDateTime:                       s.Properties.EndDateTime,
				PrincipalName:                     *group.GetDisplayName(),
				PrincipalType:                     armauthorization.PrincipalTypeGroup,
				RoleAssignmentScheduleRequestName: *s.Name,
				RoleName:                          *roleDefinition.Properties.RoleName,
				Scope:                             *s.Properties.Scope,
				StartDateTime:                     s.Properties.StartDateTime,
			})
		}
	}

	userAssignmentSchedulesToDelete, err := role_assignment_schedule.FilterForRoleAssignmentSchedulesToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingUserRoleAssignmentSchedules,
		userAssignmentSchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range userAssignmentSchedulesToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			scope,
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
			roleAssignmentScheduleDeletes = append(roleAssignmentScheduleDeletes, &core.RoleAssignmentScheduleDelete{
				Cancel:        false,
				EndDateTime:   s.Properties.EndDateTime,
				PrincipalName: *user.GetUserPrincipalName(),
				PrincipalType: armauthorization.PrincipalTypeUser,
				RoleAssignmentScheduleRequest: &armauthorization.RoleAssignmentScheduleRequest{
					Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
						Justification:                  to.Ptr("Managed by Sheriff"),
						PrincipalID:                    s.Properties.PrincipalID,
						RequestType:                    to.Ptr(armauthorization.RequestTypeAdminRemove),
						RoleDefinitionID:               s.Properties.RoleDefinitionID,
						TargetRoleAssignmentScheduleID: s.ID,
					},
				},
				RoleAssignmentScheduleRequestName: uuid.New().String(),
				RoleName:                          *roleDefinition.Properties.RoleName,
				Scope:                             *s.Properties.Scope,
				StartDateTime:                     s.Properties.StartDateTime,
			})
		} else {
			roleAssignmentScheduleDeletes = append(roleAssignmentScheduleDeletes, &core.RoleAssignmentScheduleDelete{
				Cancel:                            true,
				EndDateTime:                       s.Properties.EndDateTime,
				PrincipalName:                     *user.GetUserPrincipalName(),
				PrincipalType:                     armauthorization.PrincipalTypeUser,
				RoleAssignmentScheduleRequestName: *s.Name,
				RoleName:                          *roleDefinition.Properties.RoleName,
				Scope:                             *s.Properties.Scope,
				StartDateTime:                     s.Properties.StartDateTime,
			})
		}
	}

	return roleAssignmentScheduleDeletes, nil
}
