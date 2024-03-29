package role_management_policy_update

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/ahmetb/go-linq/v3"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-test/deep"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy"
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy_assignment"
	"github.com/gofrontier-com/sheriff/pkg/util/role_management_policy_classification_rule"
)

func GetRoleManagementPolicyUpdates(
	clientFactory *armauthorization.ClientFactory,
	defaultRoleManagementPolicyPropertiesData string,
	config *core.AzureRmConfig,
	subscriptionId string,
) ([]*core.RoleManagementPolicyUpdate, error) {
	var roleManagementPolicyUpdates []*core.RoleManagementPolicyUpdate

	scopeRoleNameCombinations := config.GetScopeRoleNameCombinations(subscriptionId)

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

		policy := config.GetPolicyByRoleName(c.RoleName)

		var roleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset
		if policy != nil {
			roleManagementPolicyRulesetReferences := policy.GetRulesetReferencesForScope(c.Scope, subscriptionId)

			var rulesetNames []string
			linq.From(roleManagementPolicyRulesetReferences).SelectT(func(r *core.RulesetReference) string {
				return r.RulesetName
			}).ToSlice(&rulesetNames)

			linq.From(config.Rulesets).WhereT(func(s *core.RoleManagementPolicyRuleset) bool {
				return linq.From(rulesetNames).Contains(s.Name)
			}).ToSlice(&roleManagementPolicyRulesets)

			if len(roleManagementPolicyRulesets) != len(rulesetNames) {
				panic("ruleset count does not match ruleset reference count")
			}
		} else {
			roleManagementPolicyRulesets = []*core.RoleManagementPolicyRuleset{}
		}

		for _, roleManagementPolicyRuleset := range roleManagementPolicyRulesets {
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
