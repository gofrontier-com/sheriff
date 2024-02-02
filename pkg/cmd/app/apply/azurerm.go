package apply

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/go-test/deep"
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/azurerm_config"
	"github.com/gofrontier-com/sheriff/pkg/util/role_assignment_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/role_assignment_schedule_create"
	"github.com/gofrontier-com/sheriff/pkg/util/role_assignment_schedule_delete"
	"github.com/gofrontier-com/sheriff/pkg/util/role_assignment_schedule_update"
	"github.com/gofrontier-com/sheriff/pkg/util/role_definition"
	"github.com/gofrontier-com/sheriff/pkg/util/role_eligibility_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/role_eligibility_schedule_create"
	"github.com/gofrontier-com/sheriff/pkg/util/role_eligibility_schedule_delete"
	"github.com/gofrontier-com/sheriff/pkg/util/role_eligibility_schedule_update"
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy_update"
	"github.com/golang-jwt/jwt/v5"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

//go:embed default_role_management_policy.json
var defaultRoleManagementPolicyPropertiesData string

var (
	requiredActionsToApply = []string{
		"*",
		"Microsoft.Authorization/*",
	}
	requiredActionsToPlan = []string{
		"*/read",
		"Microsoft.Authorization/*/read",
	}
)

func init() {
	deep.NilSlicesAreEmpty = true
}

func ApplyAzureRm(configDir string, subscriptionId string, planOnly bool) error {
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionId)

	var warnings []string

	output.PrintlnInfo("Initialising...")

	output.PrintlnInfo("- Loading and validating config")

	config, err := azurerm_config.Load(configDir)
	if err != nil {
		if _, ok := err.(*core.ConfigurationEmptyError); ok {
			warnings = append(warnings, "Configuration is empty, is the config path correct?")
		} else {
			return err
		}
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	output.PrintlnfInfo("- Authenticating to Azure Management and Microsoft Graph APIs")

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

	var requiredActions []string
	if planOnly {
		requiredActions = requiredActionsToPlan
	} else {
		requiredActions = requiredActionsToApply
	}
	err = checkPermissions(clientFactory, credential, scope, requiredActions)
	if err != nil {
		return err
	}

	if len(warnings) > 0 {
		output.PrintlnWarn("!!! One or more warnings were generated !!!")

		for _, w := range warnings {
			output.PrintlnfWarn("- %s", w)
		}

		output.Printf("\n")
	}

	output.PrintlnInfo("Sheriff is ready to go!\n")

	output.PrintlnfInfo("Generating plan for...")

	output.PrintlnInfo("- Active assignments")

	groupAssignmentSchedules := config.GetGroupAssignmentSchedules(subscriptionId)

	existingGroupRoleAssignmentSchedules, err := role_assignment_schedule.GetRoleAssignmentSchedules(
		clientFactory,
		scope,
		func(s *armauthorization.RoleAssignmentSchedule) bool {
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeGroup &&
				*s.Properties.AssignmentType == armauthorization.AssignmentTypeAssigned
		},
	)
	if err != nil {
		return err
	}

	userAssignmentSchedules := config.GetUserAssignmentSchedules(subscriptionId)

	existingUserRoleAssignmentSchedules, err := role_assignment_schedule.GetRoleAssignmentSchedules(
		clientFactory,
		scope,
		func(s *armauthorization.RoleAssignmentSchedule) bool {
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeUser &&
				*s.Properties.AssignmentType == armauthorization.AssignmentTypeAssigned
		},
	)
	if err != nil {
		return err
	}

	roleAssignmentScheduleCreates, err := role_assignment_schedule_create.GetRoleAssignmentScheduleCreates(
		clientFactory,
		graphServiceClient,
		scope,
		groupAssignmentSchedules,
		existingGroupRoleAssignmentSchedules,
		userAssignmentSchedules,
		existingUserRoleAssignmentSchedules,
	)
	if err != nil {
		return err
	}

	roleAssignmentScheduleUpdates, err := role_assignment_schedule_update.GetRoleAssignmentScheduleUpdates(
		clientFactory,
		graphServiceClient,
		scope,
		groupAssignmentSchedules,
		existingGroupRoleAssignmentSchedules,
		userAssignmentSchedules,
		existingUserRoleAssignmentSchedules,
	)
	if err != nil {
		return err
	}

	roleAssignmentScheduleDeletes, err := role_assignment_schedule_delete.GetRoleAssignmentScheduleDeletes(
		clientFactory,
		graphServiceClient,
		scope,
		groupAssignmentSchedules,
		existingGroupRoleAssignmentSchedules,
		userAssignmentSchedules,
		existingUserRoleAssignmentSchedules,
	)
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Eligible assignments")

	groupEligibilitySchedules := config.GetGroupEligibilitySchedules(subscriptionId)

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

	userEligibilitySchedules := config.GetUserEligibilitySchedules(subscriptionId)

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

	roleEligibilityScheduleCreates, err := role_eligibility_schedule_create.GetRoleEligibilityScheduleCreates(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibilitySchedules,
		existingGroupRoleEligibilitySchedules,
		userEligibilitySchedules,
		existingUserRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	roleEligibilityScheduleUpdates, err := role_eligibility_schedule_update.GetRoleEligibilityScheduleUpdates(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibilitySchedules,
		existingGroupRoleEligibilitySchedules,
		userEligibilitySchedules,
		existingUserRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	roleEligibilityScheduleDeletes, err := role_eligibility_schedule_delete.GetRoleEligibilityScheduleDeletes(
		clientFactory,
		graphServiceClient,
		scope,
		groupEligibilitySchedules,
		existingGroupRoleEligibilitySchedules,
		userEligibilitySchedules,
		existingUserRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Role management policies\n")

	scopeRoleNameCombinations := config.GetScopeRoleNameCombinations(subscriptionId)

	rulesetReferences := config.GetRulesetReferences(subscriptionId)

	roleManagementPolicyUpdates, err := role_management_policy_update.GetRoleManagementPolicyUpdates(
		clientFactory,
		defaultRoleManagementPolicyPropertiesData,
		scopeRoleNameCombinations,
		rulesetReferences,
		config.Rulesets,
	)
	if err != nil {
		return err
	}

	if planOnly {
		output.PrintlnInfo("Sheriff would perform the following actions:\n")
	} else {
		output.PrintlnInfo("Sheriff will perform the following actions:\n")
	}

	printPlan(
		roleAssignmentScheduleCreates,
		roleAssignmentScheduleUpdates,
		roleAssignmentScheduleDeletes,
		roleEligibilityScheduleCreates,
		roleEligibilityScheduleUpdates,
		roleEligibilityScheduleDeletes,
		roleManagementPolicyUpdates,
	)

	if planOnly {
		return nil
	}

	if len(roleAssignmentScheduleCreates)+len(roleAssignmentScheduleUpdates)+len(roleAssignmentScheduleDeletes)+len(roleEligibilityScheduleCreates)+len(roleEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates)+len(roleEligibilityScheduleDeletes) == 0 {
		output.PrintlnInfo("\nNothing to do!")
		return nil
	}

	output.PrintlnInfo("\nApplying plan...\n")

	roleManagementPoliciesClient := clientFactory.NewRoleManagementPoliciesClient()
	roleAssignmentScheduleRequestsClient := clientFactory.NewRoleAssignmentScheduleRequestsClient()
	roleEligibilityScheduleRequestsClient := clientFactory.NewRoleEligibilityScheduleRequestsClient()

	for _, u := range roleManagementPolicyUpdates {
		output.PrintlnfInfo(
			"Updating role management policy for role \"%s\" at scope \"%s\"",
			u.RoleName,
			u.Scope,
		)

		_, err = roleManagementPoliciesClient.Update(
			context.Background(),
			u.Scope,
			*u.RoleManagementPolicy.Name,
			*u.RoleManagementPolicy,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, c := range roleAssignmentScheduleCreates {
		output.PrintlnfInfo(
			"Creating active assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			c.PrincipalType,
			c.PrincipalName,
			c.RoleName,
			c.Scope,
		)

		_, err = roleAssignmentScheduleRequestsClient.Create(
			context.Background(),
			c.Scope,
			c.RoleAssignmentScheduleRequestName,
			*c.RoleAssignmentScheduleRequest,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, u := range roleAssignmentScheduleUpdates {
		output.PrintlnfInfo(
			"Updating active assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			u.PrincipalType,
			u.PrincipalName,
			u.RoleName,
			u.Scope,
		)

		_, err = roleAssignmentScheduleRequestsClient.Create(
			context.Background(),
			u.Scope,
			u.RoleAssignmentScheduleRequestName,
			*u.RoleAssignmentScheduleRequest,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, d := range roleAssignmentScheduleDeletes {
		output.PrintlnfInfo(
			"Deleting active assignment for %s \"%s\" with role \"%s\" at scope \"%s\"",
			d.PrincipalType,
			d.PrincipalName,
			d.RoleName,
			d.Scope,
		)

		if d.Cancel {
			_, err = roleAssignmentScheduleRequestsClient.Cancel(
				context.Background(),
				d.Scope,
				d.RoleAssignmentScheduleRequestName,
				nil,
			)
			if err != nil {
				return err
			}
		} else {
			_, err = roleAssignmentScheduleRequestsClient.Create(
				context.Background(),
				d.Scope,
				d.RoleAssignmentScheduleRequestName,
				*d.RoleAssignmentScheduleRequest,
				nil,
			)
			if err != nil {
				return err
			}
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

	output.PrintlnfInfo("\nApply complete: %d added, %d changed, %d deleted", len(roleAssignmentScheduleCreates)+len(roleEligibilityScheduleCreates), len(roleAssignmentScheduleUpdates)+len(roleEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates), len(roleAssignmentScheduleDeletes)+len(roleEligibilityScheduleDeletes))

	return nil
}

func checkPermissions(
	clientFactory *armauthorization.ClientFactory,
	credential *azidentity.DefaultAzureCredential,
	scope string,
	requiredActions []string,
) error {
	accessToken, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}})
	if err != nil {
		return err
	}

	token, err := jwt.Parse(accessToken.Token, nil)
	if err != nil {
		if err.Error() != "token is unverifiable: no keyfunc was provided" {
			return err
		}
	}

	principalId := token.Claims.(jwt.MapClaims)["oid"].(string)

	roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()

	hasRequiredActions := false
	roleAssignmentsClientListForScopeOptions := &armauthorization.RoleAssignmentsClientListForScopeOptions{
		Filter: to.Ptr(fmt.Sprintf("assignedTo('%s')", principalId)),
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

			thisHasRequiredActions := linq.From(roleDefinition.Properties.Permissions).SelectManyT(func(p *armauthorization.Permission) linq.Query {
				return linq.From(p.Actions)
			}).AnyWithT(func(a *string) bool {
				return linq.From(requiredActions).Contains(*a)
			})

			if thisHasRequiredActions {
				hasRequiredActions = true
				break
			}
		}
	}

	if !hasRequiredActions {
		return fmt.Errorf("authenticated principal does not have the required permissions for this action")
	}

	return nil
}

func printPlan(
	roleAssignmentScheduleCreates []*core.RoleAssignmentScheduleCreate,
	roleAssignmentScheduleUpdates []*core.RoleAssignmentScheduleUpdate,
	roleAssignmentScheduleDeletes []*core.RoleAssignmentScheduleDelete,
	roleEligibilityScheduleCreates []*core.RoleEligibilityScheduleCreate,
	roleEligibilityScheduleUpdates []*core.RoleEligibilityScheduleUpdate,
	roleEligibilityScheduleDeletes []*core.RoleEligibilityScheduleDelete,
	roleManagementPolicyUpdates []*core.RoleManagementPolicyUpdate,
) {
	builder := &strings.Builder{}

	if len(roleAssignmentScheduleCreates)+len(roleAssignmentScheduleUpdates)+len(roleAssignmentScheduleDeletes)+len(roleEligibilityScheduleCreates)+len(roleEligibilityScheduleUpdates)+len(roleEligibilityScheduleDeletes)+len(roleManagementPolicyUpdates) == 0 {
		builder.WriteString("(none)\n\n")
	} else {

		if len(roleAssignmentScheduleCreates) > 0 {
			builder.WriteString("  # Create active assignments:\n\n")
			for _, c := range roleAssignmentScheduleCreates {
				builder.WriteString(fmt.Sprintf("    + %s: %s\n", c.PrincipalType, c.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", c.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n", c.Scope))
				builder.WriteString(fmt.Sprintf("      Start: %s\n", c.StartDateTime))
				builder.WriteString(fmt.Sprintf("      End:   %s\n\n", c.EndDateTime))
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

		if len(roleAssignmentScheduleUpdates) > 0 {
			builder.WriteString("  # Update active assignments:\n\n")
			for _, u := range roleAssignmentScheduleUpdates {
				builder.WriteString(fmt.Sprintf("    ~ %s: %s\n", u.PrincipalType, u.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", u.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n", u.Scope))
				builder.WriteString(fmt.Sprintf("      Start: %s\n", u.StartDateTime))
				builder.WriteString(fmt.Sprintf("      End:   %s\n\n", u.EndDateTime))
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

		if len(roleAssignmentScheduleDeletes) > 0 {
			builder.WriteString("  # Delete active assignments:\n\n")
			for _, d := range roleAssignmentScheduleDeletes {
				builder.WriteString(fmt.Sprintf("    - %s: %s\n", d.PrincipalType, d.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:  %s\n", d.RoleName))
				builder.WriteString(fmt.Sprintf("      Scope: %s\n", d.Scope))
				builder.WriteString(fmt.Sprintf("      Start: %s\n", d.StartDateTime))
				builder.WriteString(fmt.Sprintf("      End:   %s\n\n", d.EndDateTime))
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

	builder.WriteString(fmt.Sprintf("Plan: %d to add, %d to change, %d to delete.", len(roleAssignmentScheduleCreates)+len(roleEligibilityScheduleCreates), len(roleAssignmentScheduleUpdates)+len(roleEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates), len(roleAssignmentScheduleDeletes)+len(roleEligibilityScheduleDeletes)))

	output.PrintlnInfo(builder.String())
}
