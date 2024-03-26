package group_role_management_policy_update

import (
	"encoding/json"
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-test/deep"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/unified_role_management_policy_assignment"
	kjson "github.com/microsoft/kiota-serialization-json-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupRoleManagementPolicyUpdates(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	defaultRoleManagementPolicyPropertiesData string,
	config *core.GroupsConfig,
) ([]*core.GroupRoleManagementPolicyUpdate, error) {
	// var roleManagementPolicyUpdates []*core.GroupRoleManagementPolicyUpdate

	groupNameRoleNameCombinations := config.GetGroupNameRoleNameCombinations()

	for _, c := range groupNameRoleNameCombinations {
		parser, err := kjson.NewJsonParseNode([]byte(defaultRoleManagementPolicyPropertiesData))
		if err != nil {
			panic(err)
		}

		parsedValue, err := parser.GetObjectValue(models.CreateUnifiedRoleManagementPolicyFromDiscriminatorValue)
		if err != nil {
			panic(err)
		}

		desiredRoleManagementPolicy := parsedValue.(models.UnifiedRoleManagementPolicyable)

		// slices.SortFunc(
		// 	desiredRoleManagementPolicy.GetRules(),
		// 	unified_role_management_policy_rule.SortByID,
		// )

		policy := config.GetPolicyByRoleName(c.RoleName)

		var roleManagementPolicyRulesets []*core.RoleManagementPolicyRuleset
		if policy != nil {
			roleManagementPolicyRulesetReferences := policy.GetRulesetReferencesForGroup(c.Target)

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

				rules := desiredRoleManagementPolicy.GetRules()

				ruleIndex := linq.From(rules).IndexOfT(func(s models.UnifiedRoleManagementPolicyRuleable) bool {
					return *s.GetId() == rule.ID
				})
				if ruleIndex == -1 {
					return nil, fmt.Errorf("rule with Id '%s' not found", rule.ID)
				}

				rule := rules[ruleIndex]

				ruleData, err := kjson.Marshal(rule)
				if err != nil {
					return nil, err
				}

				ruleDataStr := string(ruleData)
				ruleData = []byte(fmt.Sprintf("{%s}", ruleDataStr))

				patchedRuleData, err := jsonpatch.MergePatch(ruleData, rulePatchData)
				if err != nil {
					return nil, err
				}

				patchedRuleParser, err := kjson.NewJsonParseNode([]byte(patchedRuleData))
				if err != nil {
					panic(err)
				}

				patchedRuleParserValue, err := patchedRuleParser.GetObjectValue(models.CreateUnifiedRoleManagementPolicyRuleFromDiscriminatorValue)
				if err != nil {
					panic(err)
				}

				patchedRule := patchedRuleParserValue.(*models.UnifiedRoleManagementPolicyRule)

				rules[ruleIndex] = patchedRule

				desiredRoleManagementPolicy.SetRules(rules)

				// ruleType := rule.GetAdditionalData()["ruleType"].(*string)

				// switch *ruleType {
				// case "RoleManagementPolicyApprovalRule":
				// 	rule := rule.(*models.UnifiedRoleManagementPolicyApprovalRule)

				// 	ruleData, err := kjson.Marshal(rule)
				// 	if err != nil {
				// 		return nil, err
				// 	}
				// 	_ = ruleData
				// case "RoleManagementPolicyAuthenticationContextRule":
				// 	rule := rule.(*models.UnifiedRoleManagementPolicyAuthenticationContextRule)
				// 	_ = rule
				// case "RoleManagementPolicyEnablementRule":
				// 	rule := rule.(*models.UnifiedRoleManagementPolicyEnablementRule)
				// 	_ = rule
				// case "RoleManagementPolicyExpirationRule":
				// 	rule := rule.(*models.UnifiedRoleManagementPolicyExpirationRule)
				// 	_ = rule
				// case "RoleManagementPolicyNotificationRule":
				// 	rule := rule.(*models.UnifiedRoleManagementPolicyNotificationRule)
				// 	_ = rule

				// default:
				// 	return nil, fmt.Errorf("unknown rule type '%s'", *ruleType)
				// }
			}
		}

		group, err := group.GetGroupByName(graphServiceClient, c.Target)
		if err != nil {
			return nil, err
		}

		roleManagementPolicyAssignment, err := unified_role_management_policy_assignment.GetUnifiedRoleManagementPolicyAssignmentByGroupAndRole(
			graphServiceClient,
			*group.GetId(),
			c.RoleName,
		)
		if err != nil {
			return nil, err
		}

		// slices.SortFunc(
		// 	roleManagementPolicyAssignment.GetPolicy().GetEffectiveRules(),
		// 	unified_role_management_policy_rule.SortByID,
		// )

		roleManagementPolicy := roleManagementPolicyAssignment.GetPolicy()
		_ = roleManagementPolicy
		// roleManagementPolicy.GetBackingStore().SetInitializationCompleted(true)

		diff := deep.Equal(roleManagementPolicyAssignment.GetPolicy().GetEffectiveRules(), desiredRoleManagementPolicy.GetRules())
		_ = diff

		// if len(diff) > 0 {

		// }
		// desiredRoleManagementPolicy.GetBackingStore().SetReturnOnlyChangedValues(true)

		// desiredRoleManagementPolicy.

		// things := desiredRoleManagementPolicy.GetBackingStore().Enumerate()
		// foo := things
		// _ = foo
	}

	return nil, nil
}
