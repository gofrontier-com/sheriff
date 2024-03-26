package role_management_policy_assignment

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

// TODO: Fix this name to include scope.
func GetRoleManagementPolicyAssignmentByRole(clientFactory *armauthorization.ClientFactory, scope string, roleName string) (*armauthorization.RoleManagementPolicyAssignment, error) {
	// 	roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, scope, roleName)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	accessToken, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}})
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	token, err := jwt.Parse(accessToken.Token, nil)
	// 	if err != nil {
	// 		if err.Error() != "token is unverifiable: no keyfunc was provided" {
	// 			return nil, err
	// 		}
	// 	}

	// 	client := &http.Client{}

	// 	request, err := http.NewRequest("GET", fmt.Sprintf("https://management.azure.com%s/providers/Microsoft.Authorization/roleManagementPolicyAssignments", scope), nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Raw))

	// 	query := request.URL.Query()
	// 	query.Add("api-version", "2020-10-01")
	// 	query.Add("$filter", fmt.Sprintf("roleDefinitionId eq '%s'", *roleDefinition.ID))
	// 	request.URL.RawQuery = query.Encode()

	// 	response, err := client.Do(request)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer response.Body.Close()

	// 	body, err := io.ReadAll(response.Body)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	_ = body

	// 	return nil, nil

	roleManagementPolicyAssignments, err := GetRoleManagementPolicyAssignments(
		clientFactory,
		scope,
		func(r *armauthorization.RoleManagementPolicyAssignment) bool {
			return *r.Properties.PolicyAssignmentProperties.RoleDefinition.DisplayName == roleName
		},
	)
	if err != nil {
		return nil, err
	}

	if len(roleManagementPolicyAssignments) == 0 {
		return nil, fmt.Errorf("role management policy assignment at scope \"%s\" for role name \"%s\" not found", scope, roleName)
	}

	if len(roleManagementPolicyAssignments) > 1 {
		return nil, fmt.Errorf("multiple role management policy assignments at scope \"%s\" for role name \"%s\" found", scope, roleName)
	}

	return roleManagementPolicyAssignments[0], nil
}
