package role_management_policy_classification_rule

import (
	"cmp"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func SortByID(a, b armauthorization.RoleManagementPolicyRuleClassification) int {
	return cmp.Compare(*a.GetRoleManagementPolicyRule().ID, *b.GetRoleManagementPolicyRule().ID)
}
