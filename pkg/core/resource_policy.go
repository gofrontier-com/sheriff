package core

import (
	"fmt"

	"github.com/gofrontier-com/sheriff/pkg/util/regexp_patterns"
)

func (p *ResourcePolicy) GetRulesetReferencesForScope(scope string, subscriptionId string) []*RulesetReference {
	if regexp_patterns.SubscriptionId.MatchString(scope) {
		if p.Subscription != nil {
			return p.Subscription
		} else {
			return p.Default
		}
	} else if regexp_patterns.ResourceGroupId.MatchString(scope) {
		groups := regexp_patterns.ResourceGroupId.FindStringSubmatch(scope)
		if p.ResourceGroups[groups[1]] != nil {
			return p.ResourceGroups[groups[1]]
		} else {
			return p.Default
		}
	} else if regexp_patterns.ResourceId.MatchString(scope) {
		groups := regexp_patterns.ResourceId.FindStringSubmatch(scope)
		if p.Resources[groups[1]] != nil {
			return p.Resources[groups[1]]
		} else {
			return p.Default
		}
	}

	panic(fmt.Sprintf("scope '%s' is not valid", scope))
}
