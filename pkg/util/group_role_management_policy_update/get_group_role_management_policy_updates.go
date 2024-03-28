package group_role_management_policy_update

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/unified_role_management_policy_assignment"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	kjson "github.com/microsoft/kiota-serialization-json-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// var expirationAdminAssignmentData string

func GetGroupRoleManagementPolicyUpdates(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	defaultRoleManagementPolicyPropertiesData string,
	config *core.GroupsConfig,
) ([]*core.GroupRoleManagementPolicyUpdate, error) {
	var roleManagementPolicyUpdates []*core.GroupRoleManagementPolicyUpdate

	groupNameRoleNameCombinations := config.GetGroupNameRoleNameCombinations()

	for _, c := range groupNameRoleNameCombinations {
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

		roleManagementPolicyParser, err := kjson.NewJsonParseNode([]byte(defaultRoleManagementPolicyPropertiesData))
		if err != nil {
			panic(err)
		}

		roleManagementPolicyParsedValue, err := roleManagementPolicyParser.GetObjectValue(models.CreateUnifiedRoleManagementPolicyFromDiscriminatorValue)
		if err != nil {
			panic(err)
		}

		roleManagementPolicy := roleManagementPolicyParsedValue.(models.UnifiedRoleManagementPolicyable)

		rules := roleManagementPolicy.GetRules()

		for _, roleManagementPolicyRuleset := range roleManagementPolicyRulesets {
			for _, rulesetRule := range roleManagementPolicyRuleset.Rules {
				if rulesetRule.Patch == nil {
					continue
				}

				ruleIndex := linq.From(rules).IndexOfT(func(s models.UnifiedRoleManagementPolicyRuleable) bool {
					return *s.GetId() == rulesetRule.ID
				})
				if ruleIndex == -1 {
					return nil, fmt.Errorf("rule with Id '%s' not found", rulesetRule.ID)
				}

				rule := rules[ruleIndex]

				ruleType := rule.GetAdditionalData()["ruleType"].(*string)

				switch *ruleType {
				case "RoleManagementPolicyApprovalRule":
				case "RoleManagementPolicyAuthenticationContextRule":
				case "RoleManagementPolicyEnablementRule":
				case "RoleManagementPolicyExpirationRule":
					if rulesetRule.Patch.(map[string]interface{})["isExpirationRequired"] != nil {
						isExpirationRequired := rulesetRule.Patch.(map[string]interface{})["isExpirationRequired"].(bool)
						rule.(*models.UnifiedRoleManagementPolicyExpirationRule).SetIsExpirationRequired(&isExpirationRequired)
					}
					if rulesetRule.Patch.(map[string]interface{})["maximumDuration"] != nil {
						duration, err := serialization.ParseISODuration(rulesetRule.Patch.(map[string]interface{})["maximumDuration"].(string))
						if err != nil {
							panic(err)
						}
						rule.(*models.UnifiedRoleManagementPolicyExpirationRule).SetMaximumDuration(duration)
					}
				case "RoleManagementPolicyNotificationRule":
				default:
					return nil, fmt.Errorf("unknown rule type '%s'", *ruleType)
				}
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

		existingRoleManagementPolicy := roleManagementPolicyAssignment.GetPolicy()
		existingRules := existingRoleManagementPolicy.GetEffectiveRules()

		changes := []string{}
		for _, existingRule := range existingRules {
			ruleType := existingRule.GetOdataType()

			switch *ruleType {
			case "#microsoft.graph.unifiedRoleManagementPolicyApprovalRule":
			case "#microsoft.graph.unifiedRoleManagementPolicyAuthenticationContextRule":
			case "#microsoft.graph.unifiedRoleManagementPolicyEnablementRule":
			case "#microsoft.graph.unifiedRoleManagementPolicyExpirationRule":
				rule := linq.From(rules).
					SingleWithT(func(r models.UnifiedRoleManagementPolicyRuleable) bool {
						return *r.GetId() == *existingRule.GetId()
					}).(*models.UnifiedRoleManagementPolicyExpirationRule)

				if *rule.GetIsExpirationRequired() != *existingRule.(*models.UnifiedRoleManagementPolicyExpirationRule).GetIsExpirationRequired() {
					changes = append(changes, fmt.Sprintf("%s/isExpirationRequired (%t -> %t)", *existingRule.GetId(), *existingRule.(*models.UnifiedRoleManagementPolicyExpirationRule).GetIsExpirationRequired(), *rule.GetIsExpirationRequired()))
					existingRule.(*models.UnifiedRoleManagementPolicyExpirationRule).SetIsExpirationRequired(rule.GetIsExpirationRequired())
				}

				if *rule.GetMaximumDuration() != *existingRule.(*models.UnifiedRoleManagementPolicyExpirationRule).GetMaximumDuration() {
					existingRule.(*models.UnifiedRoleManagementPolicyExpirationRule).SetMaximumDuration(rule.GetMaximumDuration())
					changes = append(changes, *existingRule.GetId())
				}
			case "#microsoft.graph.unifiedRoleManagementPolicyNotificationRule":
			default:
				return nil, fmt.Errorf("unknown rule type '%s'", *ruleType)
			}
		}

		if len(changes) > 0 {
			existingRoleManagementPolicy.SetRules(existingRules)
			roleManagementPolicyUpdates = append(roleManagementPolicyUpdates, &core.GroupRoleManagementPolicyUpdate{
				Changes:              changes,
				ManagedGroupName:     c.Target,
				RoleManagementPolicy: existingRoleManagementPolicy,
				RoleName:             c.RoleName,
			})
		}
	}

	return roleManagementPolicyUpdates, nil
}
