package unified_role_management_policy_assignment

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/policies"
)

func GetUnifiedRoleManagementPolicyAssignmentByGroupAndRole(graphServiceClient *msgraphsdkgo.GraphServiceClient, groupId string, roleName string) (models.UnifiedRoleManagementPolicyAssignmentable, error) {
	requestConfiguration := &policies.RoleManagementPolicyAssignmentsRequestBuilderGetRequestConfiguration{
		QueryParameters: &policies.RoleManagementPolicyAssignmentsRequestBuilderGetQueryParameters{
			Expand: []string{"policy($expand=effectiveRules)"},
			Filter: to.Ptr(fmt.Sprintf("scopeId eq '%s' and scopeType eq 'Group' and roleDefinitionId eq '%s'", groupId, strings.ToLower(roleName))),
		},
	}
	result, err := graphServiceClient.Policies().RoleManagementPolicyAssignments().Get(context.Background(), requestConfiguration)
	if err != nil {
		return nil, err
	}

	unifiedRoleManagementPolicyAssignments := result.GetValue()

	if len(unifiedRoleManagementPolicyAssignments) == 0 {
		return nil, fmt.Errorf("unified role management policy assignment for group Id \"%s\" and role name \"%s\" not found", groupId, roleName)
	}

	if len(unifiedRoleManagementPolicyAssignments) > 1 {
		return nil, fmt.Errorf("multiple unified role management policy assignments for group Id \"%s\" and role name \"%s\" found", groupId, roleName)
	}

	return unifiedRoleManagementPolicyAssignments[0], nil
}
