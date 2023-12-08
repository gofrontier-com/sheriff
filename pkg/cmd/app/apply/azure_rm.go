package apply

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/frontierdigital/sheriff/pkg/core"
	"github.com/frontierdigital/sheriff/pkg/util/azure_rm_config"
	"github.com/frontierdigital/sheriff/pkg/util/filter"
	"github.com/frontierdigital/sheriff/pkg/util/group"
	"github.com/frontierdigital/sheriff/pkg/util/role_assignment"
	"github.com/frontierdigital/sheriff/pkg/util/role_definition"
	"github.com/frontierdigital/sheriff/pkg/util/role_eligibility_schedule"
	"github.com/frontierdigital/sheriff/pkg/util/user"
	"github.com/frontierdigital/utils/output"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/exp/slices"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func ApplyAzureRm(configDir string, subscriptionId string, dryRun bool) error {
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionId)

	PrintHeader(configDir, scope, dryRun)

	output.PrintlnfInfo("Loading and validating Azure RM config from %s", configDir)

	config, err := azure_rm_config.Load(configDir)
	if err != nil {
		return err
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	output.PrintlnfInfo("Authenticating to the Azure Management API and checking for necessary permissions")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	clientFactory, err := armauthorization.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return err
	}

	graphServiceClient, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return err
	}

	CheckPermissions(clientFactory, graphServiceClient, scope)

	output.PrintlnfInfo("Comparing config against scope \"%s\"", scope)

	groupActiveAssignments := config.GetGroupActiveAssignments(subscriptionId)

	existingGroupRoleAssignments, err := role_assignment.GetRoleAssignments(
		clientFactory,
		cache,
		scope,
		func(a *armauthorization.RoleAssignment) bool {
			return *a.Properties.PrincipalType == armauthorization.PrincipalTypeGroup
		},
	)
	if err != nil {
		return err
	}

	userActiveAssignments := config.GetUserActiveAssignments(subscriptionId)

	existingUserRoleAssignments, err := role_assignment.GetRoleAssignments(
		clientFactory,
		cache,
		scope,
		func(a *armauthorization.RoleAssignment) bool {
			return *a.Properties.PrincipalType == armauthorization.PrincipalTypeUser
		},
	)
	if err != nil {
		return err
	}

	roleAssignmentCreates, err := getRoleAssignmentCreates(
		clientFactory,
		graphServiceClient,
		scope,
		groupActiveAssignments,
		existingGroupRoleAssignments,
		userActiveAssignments,
		existingUserRoleAssignments,
	)
	if err != nil {
		return err
	}

	roleAssignmentDeletes, err := getRoleAssignmentDeletes(
		clientFactory,
		graphServiceClient,
		scope,
		groupActiveAssignments,
		existingGroupRoleAssignments,
		userActiveAssignments,
		existingUserRoleAssignments,
	)
	if err != nil {
		return err
	}

	groupEligibleAssignments := config.GetGroupEligibleAssignments(subscriptionId)

	existingGroupRoleEligibilitySchedules, err := role_eligibility_schedule.GetRoleEligibilitySchedules(
		clientFactory,
		cache,
		scope,
		func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeGroup
		},
	)
	if err != nil {
		return err
	}

	userEligibleAssignments := config.GetUserEligibleAssignments(subscriptionId)

	existingUserRoleEligibilitySchedules, err := role_eligibility_schedule.GetRoleEligibilitySchedules(
		clientFactory,
		cache,
		scope,
		func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeUser
		},
	)
	if err != nil {
		return err
	}

	roleEligibilityScheduleCreates, err := getRoleEligibilityScheduleCreates(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibleAssignments,
		existingGroupRoleEligibilitySchedules,
		userEligibleAssignments,
		existingUserRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	roleEligibilityScheduleUpdates, err := getRoleEligibilityScheduleUpdates(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibleAssignments,
		existingGroupRoleEligibilitySchedules,
		userEligibleAssignments,
		existingUserRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	roleEligibilityScheduleDeletes, err := getRoleEligibilityScheduleDeletes(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibleAssignments,
		existingGroupRoleEligibilitySchedules,
		userEligibleAssignments,
		existingUserRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	output.PrintlnInfo(strings.Repeat("-", 78))
	output.PrintlnfInfo("Active assignments to create: %d", len(roleAssignmentCreates))
	output.PrintlnfInfo("Active assignments to delete: %d", len(roleAssignmentDeletes))
	output.PrintlnfInfo("Eligible assignments to create: %d", len(roleEligibilityScheduleCreates))
	output.PrintlnfInfo("Eligible assignments to update: %d", len(roleEligibilityScheduleUpdates))
	output.PrintlnfInfo("Eligible assignments to delete: %d", len(roleEligibilityScheduleDeletes))
	output.PrintlnInfo(strings.Repeat("-", 78))

	if dryRun {
		return nil
	}

	roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()
	roleEligibilityScheduleRequestsClient := clientFactory.NewRoleEligibilityScheduleRequestsClient()

	for _, c := range roleAssignmentCreates {
		output.PrintlnfInfo(
			"Creating active assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			c.PrincipalType,
			c.PrincipalName,
			c.RoleName,
			c.Scope,
		)

		_, err = roleAssignmentsClient.Create(
			context.Background(),
			c.Scope,
			c.RoleAssignmentName,
			*c.RoleAssignmentCreateParameters,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, d := range roleAssignmentDeletes {
		output.PrintlnfInfo(
			"Deleting active assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			d.PrincipalType,
			d.PrincipalName,
			d.RoleName,
			d.Scope,
		)

		_, err = roleAssignmentsClient.DeleteByID(context.Background(), d.RoleAssignmentID, nil)
		if err != nil {
			return err
		}
	}

	for _, c := range roleEligibilityScheduleCreates {
		output.PrintlnfInfo(
			"Creating eligible assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			c.PrincipalType,
			c.PrincipalName,
			c.RoleName,
			c.Scope,
		)

		_, err = roleEligibilityScheduleRequestsClient.Create(
			context.Background(),
			c.Scope,
			c.RoleEligibilityScheduleRequestName,
			*c.RoleEligibilityScheduleRequest,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, u := range roleEligibilityScheduleUpdates {
		output.PrintlnfInfo(
			"Updating eligible assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			u.PrincipalType,
			u.PrincipalName,
			u.RoleName,
			u.Scope,
		)

		_, err = roleEligibilityScheduleRequestsClient.Create(
			context.Background(),
			u.Scope,
			u.RoleEligibilityScheduleRequestName,
			*u.RoleEligibilityScheduleRequest,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, d := range roleEligibilityScheduleDeletes {
		output.PrintlnfInfo(
			"Deleting eligible assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			d.PrincipalType,
			d.PrincipalName,
			d.RoleName,
			d.Scope,
		)

		if d.Cancel {
			_, err = roleEligibilityScheduleRequestsClient.Cancel(
				context.Background(),
				d.Scope,
				d.RoleEligibilityScheduleRequestName,
				nil,
			)
			if err != nil {
				return err
			}
		} else {
			_, err = roleEligibilityScheduleRequestsClient.Create(
				context.Background(),
				d.Scope,
				d.RoleEligibilityScheduleRequestName,
				*d.RoleEligibilityScheduleRequest,
				nil,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func filterForActiveAssignmentsToCreate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	activeAssignments []*core.ActiveAssignment,
	existingRoleAssignments []*armauthorization.RoleAssignment,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, gocache.Cache, string) (*string, error),
) (filtered []*core.ActiveAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(activeAssignments, func(a *core.ActiveAssignment) bool {
		idx := slices.IndexFunc(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				cache,
				*r.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return a.Scope == *r.Properties.Scope &&
				a.RoleName == *roleDefinition.Properties.RoleName &&
				a.PrincipalName == *principalName
		})

		return idx == -1
	})

	return
}

func filterForEligibleAssignmentsToCreate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	eligibleAssignments []*core.EligibleAssignment,
	existingRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, gocache.Cache, string) (*string, error),
) (filtered []*core.EligibleAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(eligibleAssignments, func(a *core.EligibleAssignment) bool {
		idx := slices.IndexFunc(existingRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				scope,
				*s.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				cache,
				*s.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return a.Scope == *s.Properties.Scope &&
				a.RoleName == *roleDefinition.Properties.RoleName &&
				a.PrincipalName == *principalName
		})

		return idx == -1
	})

	return
}

func filterForEligibleAssignmentsToUpdate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	eligibleAssignments []*core.EligibleAssignment,
	existingRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, gocache.Cache, string) (*string, error),
) (filtered []*core.EligibleAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(eligibleAssignments, func(a *core.EligibleAssignment) bool {
		idx := slices.IndexFunc(existingRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				scope,
				*s.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				cache,
				*s.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return a.Scope == *s.Properties.Scope &&
				a.RoleName == *roleDefinition.Properties.RoleName &&
				a.PrincipalName == *principalName
		})

		if idx == -1 {
			return false
		}

		existingRoleEligibilitySchedule := existingRoleEligibilitySchedules[idx]

		if a.StartDateTime != nil {
			if *existingRoleEligibilitySchedule.Properties.StartDateTime != *a.StartDateTime {
				return true
			}
		}

		if *existingRoleEligibilitySchedule.Properties.EndDateTime != *a.EndDateTime {
			return true
		}

		return false
	})

	return
}

func filterForRoleAssignmentsToDelete(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	existingRoleAssignments []*armauthorization.RoleAssignment,
	activeAssignments []*core.ActiveAssignment,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, gocache.Cache, string) (*string, error),
) (filtered []*armauthorization.RoleAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
		idx := slices.IndexFunc(activeAssignments, func(a *core.ActiveAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				cache,
				*r.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return *r.Properties.Scope == a.Scope &&
				*roleDefinition.Properties.RoleName == a.RoleName &&
				*principalName == a.PrincipalName
		})

		return idx == -1
	})

	return
}

func filterForRoleEligibilitySchedulesToDelete(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	existingRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	eligibleAssignments []*core.EligibleAssignment,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, gocache.Cache, string) (*string, error),
) (filtered []*armauthorization.RoleEligibilitySchedule, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(existingRoleEligibilitySchedules, func(r *armauthorization.RoleEligibilitySchedule) bool {
		idx := slices.IndexFunc(eligibleAssignments, func(a *core.EligibleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
				cache,
				*r.Properties.PrincipalID,
			)
			if err != nil {
				panic(err)
			}

			return *r.Properties.Scope == a.Scope &&
				*roleDefinition.Properties.RoleName == a.RoleName &&
				*principalName == a.PrincipalName
		})

		return idx == -1
	})

	return
}

func getRoleAssignmentCreates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupActiveAssignments []*core.ActiveAssignment,
	existingGroupRoleAssignments []*armauthorization.RoleAssignment,
	userActiveAssignments []*core.ActiveAssignment,
	existingUserRoleAssignments []*armauthorization.RoleAssignment,
) ([]*core.RoleAssignmentCreate, error) {
	var roleAssignmentCreates []*core.RoleAssignmentCreate

	groupActiveAssignmentsToCreate, err := filterForActiveAssignmentsToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		groupActiveAssignments,
		existingGroupRoleAssignments,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupActiveAssignmentsToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, cache, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, cache, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		roleAssignmentCreate := &core.RoleAssignmentCreate{
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleAssignmentCreateParameters: &armauthorization.RoleAssignmentCreateParameters{
				Properties: &armauthorization.RoleAssignmentProperties{
					Description:      to.Ptr("Managed by Sheriff"),
					PrincipalID:      group.GetId(),
					PrincipalType:    to.Ptr(armauthorization.PrincipalTypeGroup),
					RoleDefinitionID: roleDefinition.ID,
				},
			},
			RoleAssignmentName: uuid.New().String(),
			RoleName:           *roleDefinition.Properties.RoleName,
			Scope:              a.Scope,
		}
		roleAssignmentCreates = append(roleAssignmentCreates, roleAssignmentCreate)
	}

	userActiveAssignmentsToCreate, err := filterForActiveAssignmentsToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		userActiveAssignments,
		existingUserRoleAssignments,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userActiveAssignmentsToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, cache, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, cache, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		roleAssignmentCreate := &core.RoleAssignmentCreate{
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeUser,
			RoleAssignmentCreateParameters: &armauthorization.RoleAssignmentCreateParameters{
				Properties: &armauthorization.RoleAssignmentProperties{
					Description:      to.Ptr("Managed by Sheriff"),
					PrincipalID:      user.GetId(),
					PrincipalType:    to.Ptr(armauthorization.PrincipalTypeUser),
					RoleDefinitionID: roleDefinition.ID,
				},
			},
			RoleAssignmentName: uuid.New().String(),
			RoleName:           *roleDefinition.Properties.RoleName,
			Scope:              a.Scope,
		}
		roleAssignmentCreates = append(roleAssignmentCreates, roleAssignmentCreate)
	}

	return roleAssignmentCreates, nil
}

func getRoleAssignmentDeletes(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupActiveAssignments []*core.ActiveAssignment,
	existingGroupRoleAssignments []*armauthorization.RoleAssignment,
	userActiveAssignments []*core.ActiveAssignment,
	existingUserRoleAssignments []*armauthorization.RoleAssignment,
) ([]*core.RoleAssignmentDelete, error) {
	groupRoleAssignmentsToDelete, err := filterForRoleAssignmentsToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingGroupRoleAssignments,
		groupActiveAssignments,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	userRoleAssignmentsToDelete, err := filterForRoleAssignmentsToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingUserRoleAssignments,
		userActiveAssignments,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	var roleAssignmentDeletes []*core.RoleAssignmentDelete

	for _, a := range groupRoleAssignmentsToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			cache,
			scope,
			*a.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupById(graphServiceClient, cache, *a.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		roleAssignmentDelete := &core.RoleAssignmentDelete{
			PrincipalName:    *group.GetDisplayName(),
			PrincipalType:    armauthorization.PrincipalTypeGroup,
			RoleAssignmentID: *a.ID,
			RoleName:         *roleDefinition.Properties.RoleName,
			Scope:            *a.Properties.Scope,
		}
		roleAssignmentDeletes = append(roleAssignmentDeletes, roleAssignmentDelete)
	}

	for _, a := range userRoleAssignmentsToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			cache,
			scope,
			*a.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserById(graphServiceClient, cache, *a.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		roleAssignmentDelete := &core.RoleAssignmentDelete{
			PrincipalName:    *user.GetUserPrincipalName(),
			PrincipalType:    armauthorization.PrincipalTypeUser,
			RoleAssignmentID: *a.ID,
			RoleName:         *roleDefinition.Properties.RoleName,
			Scope:            *a.Properties.Scope,
		}
		roleAssignmentDeletes = append(roleAssignmentDeletes, roleAssignmentDelete)
	}

	return roleAssignmentDeletes, nil
}

func getRoleEligibilityScheduleCreates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupEligibleAssignments []*core.EligibleAssignment,
	existingGroupRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	userEligibleAssignments []*core.EligibleAssignment,
	existingUserRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleCreate, error) {
	var roleEligibilityScheduleCreates []*core.RoleEligibilityScheduleCreate

	groupEligibleAssignmentsToCreate, err := filterForEligibleAssignmentsToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibleAssignments,
		existingGroupRoleEligibilitySchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupEligibleAssignmentsToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, cache, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, cache, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		roleEligibilityScheduleCreate := &core.RoleEligibilityScheduleCreate{
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     getScheduleInfo(a.StartDateTime, a.EndDateTime),
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              a.Scope,
		}
		roleEligibilityScheduleCreates = append(roleEligibilityScheduleCreates, roleEligibilityScheduleCreate)
	}

	userEligibleAssignmentsToCreate, err := filterForEligibleAssignmentsToCreate(
		clientFactory,
		graphServiceClient,
		scope,
		userEligibleAssignments,
		existingUserRoleEligibilitySchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userEligibleAssignmentsToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, cache, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, cache, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		roleEligibilityScheduleCreate := &core.RoleEligibilityScheduleCreate{
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeUser,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     getScheduleInfo(a.StartDateTime, a.EndDateTime),
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              a.Scope,
		}
		roleEligibilityScheduleCreates = append(roleEligibilityScheduleCreates, roleEligibilityScheduleCreate)
	}

	return roleEligibilityScheduleCreates, nil
}

func getRoleEligibilityScheduleDeletes(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupEligibleAssignments []*core.EligibleAssignment,
	existingGroupRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	userEligibleAssignments []*core.EligibleAssignment,
	existingUserRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleDelete, error) {
	var roleEligibilityScheduleDeletes []*core.RoleEligibilityScheduleDelete

	groupEligibilitySchedulesToDelete, err := filterForRoleEligibilitySchedulesToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingGroupRoleEligibilitySchedules,
		groupEligibleAssignments,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range groupEligibilitySchedulesToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			cache,
			scope,
			*s.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupById(graphServiceClient, cache, *s.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		var roleEligibilityScheduleDelete *core.RoleEligibilityScheduleDelete
		if *s.Properties.Status == armauthorization.StatusProvisioned {
			roleEligibilityScheduleDelete = &core.RoleEligibilityScheduleDelete{
				Cancel:        false,
				PrincipalName: *group.GetDisplayName(),
				PrincipalType: armauthorization.PrincipalTypeGroup,
				RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
					Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
						PrincipalID:                     s.Properties.PrincipalID,
						RequestType:                     to.Ptr(armauthorization.RequestTypeAdminRemove),
						RoleDefinitionID:                s.Properties.RoleDefinitionID,
						TargetRoleEligibilityScheduleID: s.ID,
					},
				},
				RoleEligibilityScheduleRequestName: uuid.New().String(),
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
			}
		} else {
			roleEligibilityScheduleDelete = &core.RoleEligibilityScheduleDelete{
				Cancel:                             true,
				PrincipalName:                      *group.GetDisplayName(),
				PrincipalType:                      armauthorization.PrincipalTypeGroup,
				RoleEligibilityScheduleRequestName: *s.Name,
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
			}
		}
		roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, roleEligibilityScheduleDelete)
	}

	userEligibilitySchedulesToDelete, err := filterForRoleEligibilitySchedulesToDelete(
		clientFactory,
		graphServiceClient,
		scope,
		existingUserRoleEligibilitySchedules,
		userEligibleAssignments,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, s := range userEligibilitySchedulesToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			cache,
			scope,
			*s.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserById(graphServiceClient, cache, *s.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		var roleEligibilityScheduleDelete *core.RoleEligibilityScheduleDelete
		if *s.Properties.Status == armauthorization.StatusProvisioned {
			roleEligibilityScheduleDelete = &core.RoleEligibilityScheduleDelete{
				Cancel:        false,
				PrincipalName: *user.GetDisplayName(),
				PrincipalType: armauthorization.PrincipalTypeUser,
				RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
					Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
						PrincipalID:                     s.Properties.PrincipalID,
						RequestType:                     to.Ptr(armauthorization.RequestTypeAdminRemove),
						RoleDefinitionID:                s.Properties.RoleDefinitionID,
						TargetRoleEligibilityScheduleID: s.ID,
					},
				},
				RoleEligibilityScheduleRequestName: uuid.New().String(),
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
			}
		} else {
			roleEligibilityScheduleDelete = &core.RoleEligibilityScheduleDelete{
				Cancel:                             true,
				PrincipalName:                      *user.GetDisplayName(),
				PrincipalType:                      armauthorization.PrincipalTypeUser,
				RoleEligibilityScheduleRequestName: *s.Name,
				RoleName:                           *roleDefinition.Properties.RoleName,
				Scope:                              *s.Properties.Scope,
			}
		}
		roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, roleEligibilityScheduleDelete)
	}

	return roleEligibilityScheduleDeletes, nil
}

func getRoleEligibilityScheduleUpdates(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	groupEligibleAssignments []*core.EligibleAssignment,
	existingGroupRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
	userEligibleAssignments []*core.EligibleAssignment,
	existingUserRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) ([]*core.RoleEligibilityScheduleUpdate, error) {
	var roleEligibilityScheduleUpdates []*core.RoleEligibilityScheduleUpdate

	groupEligibleAssignmentsToUpdate, err := filterForEligibleAssignmentsToUpdate(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibleAssignments,
		existingGroupRoleEligibilitySchedules,
		group.GetGroupDisplayNameById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range groupEligibleAssignmentsToUpdate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(
			clientFactory,
			cache,
			scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, cache, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		existingGroupRoleEligibilityScheduleIdx := slices.IndexFunc(existingGroupRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *group.GetId()
		})
		if existingGroupRoleEligibilityScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role eligibility schedule not found")
		}

		existingGroupRoleEligibilitySchedule := existingGroupRoleEligibilitySchedules[existingGroupRoleEligibilityScheduleIdx]

		roleEligibilityScheduleUpdate := &core.RoleEligibilityScheduleUpdate{
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     getScheduleInfo(a.StartDateTime, a.EndDateTime),
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              *existingGroupRoleEligibilitySchedule.Properties.Scope,
		}
		roleEligibilityScheduleUpdates = append(roleEligibilityScheduleUpdates, roleEligibilityScheduleUpdate)
	}

	userEligibleAssignmentsToUpdate, err := filterForEligibleAssignmentsToUpdate(
		clientFactory,
		graphServiceClient,
		scope,
		userEligibleAssignments,
		existingUserRoleEligibilitySchedules,
		user.GetUserUpnById,
	)
	if err != nil {
		return nil, err
	}

	for _, a := range userEligibleAssignmentsToUpdate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(
			clientFactory,
			cache,
			scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, cache, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		existingUserRoleEligibilityScheduleIdx := slices.IndexFunc(existingUserRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.Scope == a.Scope &&
				*s.Properties.RoleDefinitionID == *roleDefinition.ID &&
				*s.Properties.PrincipalID == *user.GetId()
		})
		if existingUserRoleEligibilityScheduleIdx == -1 {
			return nil, fmt.Errorf("existing role eligibility schedule not found")
		}

		existingUserRoleEligibilitySchedule := existingUserRoleEligibilitySchedules[existingUserRoleEligibilityScheduleIdx]

		roleEligibilityScheduleUpdate := &core.RoleEligibilityScheduleUpdate{
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     getScheduleInfo(a.StartDateTime, a.EndDateTime),
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              *existingUserRoleEligibilitySchedule.Properties.Scope,
		}
		roleEligibilityScheduleUpdates = append(roleEligibilityScheduleUpdates, roleEligibilityScheduleUpdate)
	}

	return roleEligibilityScheduleUpdates, nil
}

func getScheduleInfo(
	startDateTime *time.Time,
	endDateTime *time.Time,
) *armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo {
	var expiration armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration

	if endDateTime == nil {
		expiration = armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration{
			Type: to.Ptr(armauthorization.TypeNoExpiration),
		}
	} else {
		expiration = armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration{
			EndDateTime: endDateTime,
			Type:        to.Ptr(armauthorization.TypeAfterDateTime),
		}
	}

	if startDateTime == nil {
		return &armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo{
			Expiration:    &expiration,
			StartDateTime: to.Ptr(time.Now()),
		}
	} else {
		return &armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo{
			Expiration:    &expiration,
			StartDateTime: startDateTime,
		}
	}
}

func CheckPermissions(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
) error {
	me, err := graphServiceClient.Me().Get(context.Background(), nil)
	if err != nil {
		return err
	}

	roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()

	hasRoleAssignmentsWritePermission := false
	roleAssignmentsClientListForScopeOptions := &armauthorization.RoleAssignmentsClientListForScopeOptions{
		Filter: to.Ptr(fmt.Sprintf("assignedTo('%s')", *me.GetId())),
	}
	pager := roleAssignmentsClient.NewListForScopePager(scope, roleAssignmentsClientListForScopeOptions)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return err
		}

		for _, r := range page.Value {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				return err
			}

			for _, p := range roleDefinition.Properties.Permissions {
				for _, a := range p.Actions {
					if *a == "*" {
						hasRoleAssignmentsWritePermission = true
						break
					}
					if *a == "Microsoft.Authorization/roleAssignments/write" {
						hasRoleAssignmentsWritePermission = true
						break
					}
				}
			}
		}
	}

	if !hasRoleAssignmentsWritePermission {
		return fmt.Errorf("user does not have permission to create role assignments") // TODO: Update
	}

	return nil
}

func PrintHeader(configDir string, scope string, dryRun bool) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 78)))
	builder.WriteString(fmt.Sprintf("Config directory  | %s\n", configDir))
	builder.WriteString(fmt.Sprintf("Scope             | %s\n", scope))
	builder.WriteString(fmt.Sprintf("Mode              | %s\n", "Groups only"))
	builder.WriteString(fmt.Sprintf("Dry-run           | %v\n", dryRun))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 78)))
	output.Println(builder.String())
}

func PrintPlan(createCount int, deleteCount int, ignoreCount int) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("-", 78)))
	builder.WriteString(fmt.Sprintf("|%s|\n", strings.Repeat(" ", 76)))
	builder.WriteString(fmt.Sprintf("|         Create: %d%sDelete: %d%sIgnore: %d          |\n", createCount, strings.Repeat(" ", 15), deleteCount, strings.Repeat(" ", 15), ignoreCount))
	builder.WriteString(fmt.Sprintf("|%s|\n", strings.Repeat(" ", 76)))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 78)))
	output.Println(builder.String())
}
