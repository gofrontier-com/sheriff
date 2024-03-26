package unified_role_management_policy_rule

import (
	"cmp"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func SortByID(a, b models.UnifiedRoleManagementPolicyRuleable) int {
	return cmp.Compare(*a.GetId(), *b.GetId())
}
