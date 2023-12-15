package apply

import (
	"context"
	_ "embed"
	"encoding/json"
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
	"github.com/frontierdigital/sheriff/pkg/util/role_management_policy"
	"github.com/frontierdigital/sheriff/pkg/util/role_management_policy_assignment"
	"github.com/frontierdigital/sheriff/pkg/util/user"
	"github.com/frontierdigital/utils/output"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/exp/slices"
)

//go:embed default_role_management_policy.json
var defaultRoleManagementPolicyRulesetData string

var defaultRoleManagementPolicyRuleset *core.RoleManagementPolicyRuleset

func init() {
	roleManagementPolicyProperties := armauthorization.RoleManagementPolicyProperties{}
	err := roleManagementPolicyProperties.UnmarshalJSON([]byte(defaultRoleManagementPolicyRulesetData))
	if err != nil {
		panic(err)
	}
	defaultRoleManagementPolicyRuleset = &core.RoleManagementPolicyRuleset{
		Name:  "default",
		Rules: roleManagementPolicyProperties.Rules,
	}
}

func ApplyAzureRm(configDir string, subscriptionId string, planOnly bool) error {
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionId)

	PrintHeader(configDir, scope, planOnly)

	output.PrintlnInfo("Initialising...")

	output.PrintlnfInfo("- Loading config from %s", configDir)

	config, err := azure_rm_config.Load(configDir)
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Validating config")

	err = config.Validate()
	if err != nil {
		return err
	}

	output.PrintlnfInfo("- Authenticating to the Azure Management API")

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	clientFactory, err := armauthorization.NewClientFactory(subscriptionId, credential, nil)
	if err != nil {
		return err
	}

	graphServiceClient, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Checking for necessary permissions\n")

	CheckPermissions(clientFactory, graphServiceClient, scope)

	output.PrintlnInfo("Sheriff is ready to go!\n")

	output.PrintlnfInfo("Generating plan for...")

	output.PrintlnInfo("- Active assignments")

	groupActiveAssignments := config.GetGroupActiveAssignments(subscriptionId)

	existingGroupRoleAssignments, err := role_assignment.GetRoleAssignments(
		clientFactory,
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

	output.PrintlnInfo("- Eligible assignments")

	groupEligibleAssignments := config.GetGroupEligibleAssignments(subscriptionId)

	existingGroupRoleEligibilitySchedules, err := role_eligibility_schedule.GetRoleEligibilitySchedules(
		clientFactory,
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

	output.PrintlnInfo("- Role management policies\n")

	roleManagementPolicyUpdates, err := getRoleManagementPolicyUpdates(
		clientFactory,
		config.RoleManagementPolicyRulesets,
		append(groupEligibleAssignments, userEligibleAssignments...),
	)
	if err != nil {
		return err
	}

	if planOnly {
		output.PrintlnInfo("Sheriff would perform the following actions:\n")
	} else {
		output.PrintlnInfo("Sheriff will perform the following actions:\n")
	}

	PrintPlan(
		roleAssignmentCreates,
		roleAssignmentDeletes,
		roleEligibilityScheduleCreates,
		roleEligibilityScheduleUpdates,
		roleEligibilityScheduleDeletes,
		roleManagementPolicyUpdates,
	)

	if planOnly {
		return nil
	}

	if len(roleAssignmentCreates)+len(roleAssignmentDeletes)+len(roleEligibilityScheduleCreates)+len(roleEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates)+len(roleEligibilityScheduleDeletes) == 0 {
		output.PrintlnInfo("Nothing further to do!")
		return nil
	}

	output.PrintlnInfo("\nApplying plan...\n")

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

	output.PrintlnfInfo("\nApply complete: %d added, %d changed, %d deleted", len(roleAssignmentCreates)+len(roleEligibilityScheduleCreates), len(roleEligibilityScheduleUpdates), len(roleAssignmentDeletes)+len(roleEligibilityScheduleDeletes))

	return nil
}

func filterForActiveAssignmentsToCreate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	scope string,
	activeAssignments []*core.ActiveAssignment,
	existingRoleAssignments []*armauthorization.RoleAssignment,
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
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
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
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
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
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
				scope,
				*s.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
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
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
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
				scope,
				*s.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
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

		// if a.RoleManagementPolicyName != nil {
		// 	get_role_management_policy_assignment.GetRoleManagementPolicyAssignmentByScopeAndRole(
		// 		clientFactory,
		// 		scope,
		// 	)
		// }

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
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
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
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
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
	getPrincipalName func(*msgraphsdkgo.GraphServiceClient, string) (*string, error),
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
				scope,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			principalName, err := getPrincipalName(
				graphServiceClient,
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
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		roleAssignmentCreates = append(roleAssignmentCreates, &core.RoleAssignmentCreate{
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
		})
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
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		roleAssignmentCreates = append(roleAssignmentCreates, &core.RoleAssignmentCreate{
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
		})
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
			scope,
			*a.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupById(graphServiceClient, *a.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		roleAssignmentDeletes = append(roleAssignmentDeletes, &core.RoleAssignmentDelete{
			PrincipalName:    *group.GetDisplayName(),
			PrincipalType:    armauthorization.PrincipalTypeGroup,
			RoleAssignmentID: *a.ID,
			RoleName:         *roleDefinition.Properties.RoleName,
			Scope:            *a.Properties.Scope,
		})
	}

	for _, a := range userRoleAssignmentsToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			scope,
			*a.Properties.RoleDefinitionID,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserById(graphServiceClient, *a.Properties.PrincipalID)
		if err != nil {
			return nil, err
		}

		roleAssignmentDeletes = append(roleAssignmentDeletes, &core.RoleAssignmentDelete{
			PrincipalName:    *user.GetUserPrincipalName(),
			PrincipalType:    armauthorization.PrincipalTypeUser,
			RoleAssignmentID: *a.ID,
			RoleName:         *roleDefinition.Properties.RoleName,
			Scope:            *a.Properties.Scope,
		})
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
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		scheduleInfo := getScheduleInfo(a.StartDateTime, a.EndDateTime)
		roleEligibilityScheduleCreates = append(roleEligibilityScheduleCreates, &core.RoleEligibilityScheduleCreate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              a.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
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
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, a.RoleName)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
		if err != nil {
			return nil, err
		}

		scheduleInfo := getScheduleInfo(a.StartDateTime, a.EndDateTime)
		roleEligibilityScheduleCreates = append(roleEligibilityScheduleCreates, &core.RoleEligibilityScheduleCreate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeUser,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              a.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
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
			roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, &core.RoleEligibilityScheduleDelete{
				Cancel:        false,
				EndDateTime:   s.Properties.EndDateTime,
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
			roleEligibilityScheduleDeletes = append(roleEligibilityScheduleDeletes, &core.RoleEligibilityScheduleDelete{
				Cancel:        false,
				EndDateTime:   s.Properties.EndDateTime,
				PrincipalName: *user.GetUserPrincipalName(),
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
			scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		group, err := group.GetGroupByName(graphServiceClient, a.PrincipalName)
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

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingGroupRoleEligibilitySchedule.Properties.StartDateTime
		}

		scheduleInfo := getScheduleInfo(startTime, a.EndDateTime)
		roleEligibilityScheduleUpdates = append(roleEligibilityScheduleUpdates, &core.RoleEligibilityScheduleUpdate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *group.GetDisplayName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      group.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              *existingGroupRoleEligibilitySchedule.Properties.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
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
			scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		user, err := user.GetUserByUpn(graphServiceClient, a.PrincipalName)
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

		var startTime *time.Time
		if a.StartDateTime != nil {
			startTime = a.StartDateTime
		} else {
			startTime = existingUserRoleEligibilitySchedule.Properties.StartDateTime
		}

		scheduleInfo := getScheduleInfo(startTime, a.EndDateTime)
		roleEligibilityScheduleUpdates = append(roleEligibilityScheduleUpdates, &core.RoleEligibilityScheduleUpdate{
			EndDateTime:   scheduleInfo.Expiration.EndDateTime,
			PrincipalName: *user.GetUserPrincipalName(),
			PrincipalType: armauthorization.PrincipalTypeGroup,
			RoleEligibilityScheduleRequest: &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:      user.GetId(),
					RequestType:      to.Ptr(armauthorization.RequestTypeAdminUpdate),
					RoleDefinitionID: roleDefinition.ID,
					ScheduleInfo:     scheduleInfo,
				},
			},
			RoleEligibilityScheduleRequestName: uuid.New().String(),
			RoleName:                           *roleDefinition.Properties.RoleName,
			Scope:                              *existingUserRoleEligibilitySchedule.Properties.Scope,
			StartDateTime:                      scheduleInfo.StartDateTime,
		})
	}

	return roleEligibilityScheduleUpdates, nil
}

func getRoleManagementPolicyUpdates(
	clientFactory *armauthorization.ClientFactory,
	roleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset,
	eligibleAssignments []*core.EligibleAssignment,
) ([]*core.RoleManagementPolicyUpdate, error) {
	var roleManagementPolicyUpdates []*core.RoleManagementPolicyUpdate

	for _, a := range eligibleAssignments {
		var roleManagementPolicyRuleset *core.RoleManagementPolicyRuleset

		if a.RoleManagementPolicyName == nil {
			roleManagementPolicyRuleset = defaultRoleManagementPolicyRuleset
		} else {
			idx := slices.IndexFunc(roleManagementPolicyRulesets, func(r *core.RoleManagementPolicyRuleset) bool {
				return r.Name == *a.RoleManagementPolicyName
			})

			if idx == -1 {
				panic(fmt.Sprintf("role management policy ruleset with name '%s' not found", *a.RoleManagementPolicyName))
			}

			roleManagementPolicyRuleset = roleManagementPolicyRulesets[idx]
		}

		roleManagementPolicyAssignment, err := role_management_policy_assignment.GetRoleManagementPolicyAssignmentByRole(
			clientFactory,
			a.Scope,
			a.RoleName,
		)
		if err != nil {
			return nil, err
		}

		roleManagementPolicyIdParts := strings.Split(*roleManagementPolicyAssignment.Properties.PolicyID, "/")
		roleManagementPolicy, err := role_management_policy.GetRoleManagementPolicyById(
			clientFactory,
			*roleManagementPolicyAssignment.Properties.Scope,
			roleManagementPolicyIdParts[len(roleManagementPolicyIdParts)-1],
		)
		if err != nil {
			return nil, err
		}

		updateRequired := false
		for _, r := range roleManagementPolicyRuleset.Rules {
			ruleId := r.GetRoleManagementPolicyRule().ID
			idx := slices.IndexFunc(roleManagementPolicy.Properties.Rules, func(s armauthorization.RoleManagementPolicyRuleClassification) bool {
				return *s.GetRoleManagementPolicyRule().ID == *ruleId
			})

			if idx == -1 {
				return nil, fmt.Errorf("rule with Id '%s' not found", *ruleId)
			}

			existingRule := roleManagementPolicy.Properties.Rules[idx]

			ruleData, err := json.Marshal(r)
			if err != nil {
				return nil, err
			}

			existingRuleData, err := json.Marshal(existingRule)
			if err != nil {
				return nil, err
			}

			if string(ruleData) != string(existingRuleData) {
				updateRequired = true
				break
			}
		}

		if updateRequired {
			roleManagementPolicyUpdates = append(roleManagementPolicyUpdates, &core.RoleManagementPolicyUpdate{
				RoleManagementPolicyId:      *roleManagementPolicy.ID,
				RoleManagementPolicyRuleset: roleManagementPolicyRuleset,
				RoleName:                    *roleManagementPolicyAssignment.Properties.PolicyAssignmentProperties.RoleDefinition.DisplayName,
				Scope:                       *roleManagementPolicyAssignment.Properties.PolicyAssignmentProperties.Scope.ID,
			})
		}
	}

	return roleManagementPolicyUpdates, nil
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

func PrintHeader(configDir string, scope string, planOnly bool) {
	var actionText string
	if planOnly {
		actionText = "Plan"
	} else {
		actionText = "Apply"
	}

	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	builder.WriteString(fmt.Sprintf("Action   | %s\n", actionText))
	builder.WriteString(fmt.Sprintf("Mode     | %s\n", "Azure RM"))
	builder.WriteString(fmt.Sprintf("Scope    | %s\n", scope))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	output.PrintlnInfo(builder.String())
}

func PrintPlan(
	roleAssignmentCreates []*core.RoleAssignmentCreate,
	roleAssignmentDeletes []*core.RoleAssignmentDelete,
	roleEligibilityScheduleCreates []*core.RoleEligibilityScheduleCreate,
	roleEligibilityScheduleUpdates []*core.RoleEligibilityScheduleUpdate,
	roleEligibilityScheduleDeletes []*core.RoleEligibilityScheduleDelete,
	roleManagementPolicyUpdates []*core.RoleManagementPolicyUpdate,
) {
	builder := &strings.Builder{}

	if len(roleAssignmentCreates)+len(roleAssignmentDeletes)+len(roleEligibilityScheduleCreates)+len(roleEligibilityScheduleUpdates)+len(roleEligibilityScheduleDeletes)+len(roleManagementPolicyUpdates) == 0 {
		builder.WriteString("(none)\n\n")
	} else {

		if len(roleAssignmentCreates) > 0 {
			builder.WriteString("  # Create active assignments:\n\n")
			for _, c := range roleAssignmentCreates {
				builder.WriteString(fmt.Sprintf("    + %s: %s\n", c.PrincipalType, c.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", c.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n\n", c.Scope))
			}
		}

		if len(roleEligibilityScheduleCreates) > 0 {
			builder.WriteString("  # Create eligible assignments:\n\n")
			for _, c := range roleEligibilityScheduleCreates {
				builder.WriteString(fmt.Sprintf("    + %s: %s\n", c.PrincipalType, c.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", c.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n", c.Scope))
				builder.WriteString(fmt.Sprintf("      Start: %s\n", c.StartDateTime))
				builder.WriteString(fmt.Sprintf("      End:   %s\n\n", c.EndDateTime))
			}
		}

		if len(roleEligibilityScheduleUpdates) > 0 {
			builder.WriteString("  # Update eligible assignments:\n\n")
			for _, u := range roleEligibilityScheduleUpdates {
				builder.WriteString(fmt.Sprintf("    ~ %s: %s\n", u.PrincipalType, u.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", u.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n", u.Scope))
				builder.WriteString(fmt.Sprintf("      Start: %s\n", u.StartDateTime))
				builder.WriteString(fmt.Sprintf("      End:   %s\n\n", u.EndDateTime))
			}
		}

		if len(roleManagementPolicyUpdates) > 0 {
			builder.WriteString("  # Update role management policies:\n\n")
			for _, u := range roleManagementPolicyUpdates {
				builder.WriteString(fmt.Sprintf("    ~ Role: %s\n", u.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n\n", u.Scope))
			}
		}

		if len(roleAssignmentDeletes) > 0 {
			builder.WriteString("  # Delete active assignments:\n\n")
			for _, d := range roleAssignmentDeletes {
				builder.WriteString(fmt.Sprintf("    - %s: %s\n", d.PrincipalType, d.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", d.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n\n", d.Scope))
			}
		}

		if len(roleEligibilityScheduleDeletes) > 0 {
			builder.WriteString("  # Delete eligible assignments:\n\n")
			for _, d := range roleEligibilityScheduleDeletes {
				builder.WriteString(fmt.Sprintf("    - %s: %s\n", d.PrincipalType, d.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", d.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n", d.Scope))
				builder.WriteString(fmt.Sprintf("      Start: %s\n", d.StartDateTime))
				builder.WriteString(fmt.Sprintf("      End:   %s\n\n", d.EndDateTime))
			}
		}
	}

	builder.WriteString(fmt.Sprintf("Plan: %d to add, %d to change, %d to delete.", len(roleAssignmentCreates)+len(roleEligibilityScheduleCreates), len(roleEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates), len(roleAssignmentDeletes)+len(roleEligibilityScheduleDeletes)))

	output.PrintlnInfo(builder.String())
}
