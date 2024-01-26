package apply

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	jsonpatch "github.com/evanphx/json-patch"
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
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy"
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy_assignment"
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy_classification_rule"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/exp/slices"
)

//go:embed default_role_management_policy.json
var defaultRoleManagementPolicyPropertiesData string

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
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeGroup
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
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeUser
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

	allSchedules := append(groupAssignmentSchedules, userAssignmentSchedules...)
	allSchedules = append(allSchedules, groupEligibilitySchedules...)
	allSchedules = append(allSchedules, userEligibilitySchedules...)

	var scopeRoleNameCombinations []*core.ScopeRoleNameCombination
	linq.From(allSchedules).SelectT(func(s *core.Schedule) *core.ScopeRoleNameCombination {
		return &core.ScopeRoleNameCombination{
			RoleName: s.RoleName,
			Scope:    s.Scope,
		}
	}).DistinctByT(func(s *core.ScopeRoleNameCombination) string {
		return fmt.Sprintf("%s:%s", s.Scope, s.RoleName)
	}).ToSlice(&scopeRoleNameCombinations)

	// TODO: Move the above.

	rulesetReferences := config.GetRulesetReferences(subscriptionId)

	roleManagementPolicyUpdates, err := getRoleManagementPolicyUpdates(
		clientFactory,
		config.Rulesets,
		scopeRoleNameCombinations,
		rulesetReferences,
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

func getRoleManagementPolicyUpdates(
	clientFactory *armauthorization.ClientFactory,
	roleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset,
	scopeRoleNameCombinations []*core.ScopeRoleNameCombination,
	rulesetReferences []*core.RulesetReference,
) ([]*core.RoleManagementPolicyUpdate, error) {
	var roleManagementPolicyUpdates []*core.RoleManagementPolicyUpdate

	var rulesetReferenceGroups []linq.Group
	linq.From(rulesetReferences).GroupByT(func(r *core.RulesetReference) core.ScopeRoleNameCombination {
		return core.ScopeRoleNameCombination{
			RoleName: r.RoleName,
			Scope:    r.Scope,
		}
	}, func(s *core.RulesetReference) string {
		return s.RulesetName
	}).ToSlice(&rulesetReferenceGroups)

	for _, c := range scopeRoleNameCombinations {
		var desiredRoleManagementPolicyProperties armauthorization.RoleManagementPolicyProperties
		err := desiredRoleManagementPolicyProperties.UnmarshalJSON([]byte(defaultRoleManagementPolicyPropertiesData))
		if err != nil {
			panic(err)
		}

		slices.SortFunc(
			desiredRoleManagementPolicyProperties.Rules,
			role_management_policy_classification_rule.SortByID,
		)

		thisRulesetReferenceGroup := linq.From(rulesetReferenceGroups).SingleWithT(func(g linq.Group) bool {
			return g.Key.(core.ScopeRoleNameCombination).RoleName == c.RoleName && g.Key.(core.ScopeRoleNameCombination).Scope == c.Scope
		})

		var thisRoleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset
		if thisRulesetReferenceGroup != nil {
			rulesetNames := thisRulesetReferenceGroup.(linq.Group).Group
			linq.From(roleManagementPolicyRulesets).WhereT(func(s *core.RoleManagementPolicyRuleset) bool {
				return linq.From(rulesetNames).Contains(s.Name)
			}).ToSlice(&thisRoleManagementPolicyRulesets)

			if len(thisRoleManagementPolicyRulesets) != len(rulesetNames) {
				panic("ruleset count does not match ruleset reference count")
			}
		} else {
			thisRoleManagementPolicyRulesets = []*core.RoleManagementPolicyRuleset{}
		}

		for _, roleManagementPolicyRuleset := range thisRoleManagementPolicyRulesets {
			// TODO: panic on rule conflicts?

			for _, rule := range roleManagementPolicyRuleset.Rules {
				if rule.Patch == nil {
					continue
				}

				rulePatchData, err := json.Marshal(rule.Patch)
				if err != nil {
					return nil, err
				}

				ruleIndex := slices.IndexFunc(desiredRoleManagementPolicyProperties.Rules, func(s armauthorization.RoleManagementPolicyRuleClassification) bool {
					return *s.GetRoleManagementPolicyRule().ID == rule.ID
				})
				ruleIndex2 := linq.From(desiredRoleManagementPolicyProperties.Rules).IndexOfT(func(s armauthorization.RoleManagementPolicyRuleClassification) bool {
					return *s.GetRoleManagementPolicyRule().ID == rule.ID
				})
				if ruleIndex != ruleIndex2 {
					panic("index mismatch")
				}
				if ruleIndex == -1 {
					return nil, fmt.Errorf("rule with Id '%s' not found", rule.ID)
				}

				rule := desiredRoleManagementPolicyProperties.Rules[ruleIndex]

				switch *rule.GetRoleManagementPolicyRule().RuleType {
				case armauthorization.RoleManagementPolicyRuleTypeRoleManagementPolicyApprovalRule:
					rule := rule.(*armauthorization.RoleManagementPolicyApprovalRule)

					ruleData, err := rule.MarshalJSON()
					if err != nil {
						return nil, err
					}

					patchedRuleData, err := jsonpatch.MergePatch(ruleData, rulePatchData)
					if err != nil {
						return nil, err
					}

					err = rule.UnmarshalJSON(patchedRuleData)
					if err != nil {
						return nil, err
					}

					desiredRoleManagementPolicyProperties.Rules[ruleIndex] = rule
				case armauthorization.RoleManagementPolicyRuleTypeRoleManagementPolicyAuthenticationContextRule:
					rule := rule.(*armauthorization.RoleManagementPolicyAuthenticationContextRule)

					ruleData, err := rule.MarshalJSON()
					if err != nil {
						return nil, err
					}

					patchedRuleData, err := jsonpatch.MergePatch(ruleData, rulePatchData)
					if err != nil {
						return nil, err
					}

					err = rule.UnmarshalJSON(patchedRuleData)
					if err != nil {
						return nil, err
					}

					desiredRoleManagementPolicyProperties.Rules[ruleIndex] = rule
				case armauthorization.RoleManagementPolicyRuleTypeRoleManagementPolicyEnablementRule:
					rule := rule.(*armauthorization.RoleManagementPolicyEnablementRule)

					ruleData, err := rule.MarshalJSON()
					if err != nil {
						return nil, err
					}

					patchedRuleData, err := jsonpatch.MergePatch(ruleData, rulePatchData)
					if err != nil {
						return nil, err
					}

					err = rule.UnmarshalJSON(patchedRuleData)
					if err != nil {
						return nil, err
					}

					desiredRoleManagementPolicyProperties.Rules[ruleIndex] = rule
				case armauthorization.RoleManagementPolicyRuleTypeRoleManagementPolicyExpirationRule:
					rule := rule.(*armauthorization.RoleManagementPolicyExpirationRule)

					ruleData, err := rule.MarshalJSON()
					if err != nil {
						return nil, err
					}

					patchedRuleData, err := jsonpatch.MergePatch(ruleData, rulePatchData)
					if err != nil {
						return nil, err
					}

					err = rule.UnmarshalJSON(patchedRuleData)
					if err != nil {
						return nil, err
					}

					desiredRoleManagementPolicyProperties.Rules[ruleIndex] = rule
				case armauthorization.RoleManagementPolicyRuleTypeRoleManagementPolicyNotificationRule:
					rule := rule.(*armauthorization.RoleManagementPolicyNotificationRule)

					ruleData, err := rule.MarshalJSON()
					if err != nil {
						return nil, err
					}

					patchedRuleData, err := jsonpatch.MergePatch(ruleData, rulePatchData)
					if err != nil {
						return nil, err
					}

					err = rule.UnmarshalJSON(patchedRuleData)
					if err != nil {
						return nil, err
					}

					desiredRoleManagementPolicyProperties.Rules[ruleIndex] = rule
				default:
					return nil, fmt.Errorf("unknown rule type '%s'", *rule.GetRoleManagementPolicyRule().RuleType)
				}
			}
		}

		roleManagementPolicyAssignment, err := role_management_policy_assignment.GetRoleManagementPolicyAssignmentByRole(
			clientFactory,
			c.Scope,
			c.RoleName,
		)
		if err != nil {
			return nil, err
		}

		slices.SortFunc(
			roleManagementPolicyAssignment.Properties.EffectiveRules,
			role_management_policy_classification_rule.SortByID,
		)

		diff := deep.Equal(roleManagementPolicyAssignment.Properties.EffectiveRules, desiredRoleManagementPolicyProperties.Rules)
		for i, d := range diff {
			if strings.HasPrefix(d, "slice[1].ClaimValue:") {
				diff = append(diff[:i], diff[i+1:]...)
				break
			}
		}
		if len(diff) > 0 {
			roleManagementPolicyIdParts := strings.Split(*roleManagementPolicyAssignment.Properties.PolicyID, "/")
			roleManagementPolicy, err := role_management_policy.GetRoleManagementPolicyById(
				clientFactory,
				*roleManagementPolicyAssignment.Properties.Scope,
				roleManagementPolicyIdParts[len(roleManagementPolicyIdParts)-1],
			)
			if err != nil {
				return nil, err
			}

			roleManagementPolicy.Properties.Rules = desiredRoleManagementPolicyProperties.Rules
			roleManagementPolicyUpdates = append(roleManagementPolicyUpdates, &core.RoleManagementPolicyUpdate{
				RoleManagementPolicy: roleManagementPolicy,
				RoleName:             *roleManagementPolicyAssignment.Properties.PolicyAssignmentProperties.RoleDefinition.DisplayName,
				Scope:                *roleManagementPolicyAssignment.Properties.PolicyAssignmentProperties.Scope.ID,
			})
		}
	}

	return roleManagementPolicyUpdates, nil
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

func PrintPlan(
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
