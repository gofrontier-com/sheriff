package core

func (p *GroupPolicy) GetRulesetReferencesForGroup(groupName string) []*RulesetReference {
	if p.ManagedGroups[groupName] != nil {
		return p.ManagedGroups[groupName]
	} else {
		return p.Default
	}
}
