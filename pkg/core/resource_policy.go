package core

import (
	"fmt"
	"regexp"
)

var (
	resourceGroupRegex = regexp.MustCompile("^/subscriptions/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/resourceGroups/([^/]+)$")
	resourceRegex      = regexp.MustCompile("^/subscriptions/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/resourceGroups/(.+)$")
	subscriptionRegex  = regexp.MustCompile("^/subscriptions/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
)

func (p *ResourcePolicy) GetRulesetReferencesForScope(scope string, subscriptionId string) []*RulesetReference {
	if subscriptionRegex.MatchString(scope) {
		if p.Subscription != nil {
			return p.Subscription
		} else {
			return p.Default
		}
	} else if resourceGroupRegex.MatchString(scope) {
		groups := resourceGroupRegex.FindStringSubmatch(scope)
		if p.ResourceGroups[groups[1]] != nil {
			return p.ResourceGroups[groups[1]]
		} else {
			return p.Default
		}
	} else if resourceRegex.MatchString(scope) {
		groups := resourceRegex.FindStringSubmatch(scope)
		if p.Resources[groups[1]] != nil {
			return p.Resources[groups[1]]
		} else {
			return p.Default
		}
	}

	panic(fmt.Sprintf("scope '%s' is not valid", scope))
}
