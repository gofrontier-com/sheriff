package groups

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/core"
	"github.com/gofrontier-com/sheriff/pkg/util/directory_object"
	"github.com/gofrontier-com/sheriff/pkg/util/group"
	"github.com/gofrontier-com/sheriff/pkg/util/group_assignment_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/group_assignment_schedule_create"
	"github.com/gofrontier-com/sheriff/pkg/util/group_assignment_schedule_delete"
	"github.com/gofrontier-com/sheriff/pkg/util/group_assignment_schedule_update"
	"github.com/gofrontier-com/sheriff/pkg/util/group_eligibility_schedule"
	"github.com/gofrontier-com/sheriff/pkg/util/group_eligibility_schedule_create"
	"github.com/gofrontier-com/sheriff/pkg/util/group_eligibility_schedule_delete"
	"github.com/gofrontier-com/sheriff/pkg/util/group_eligibility_schedule_update"
	"github.com/gofrontier-com/sheriff/pkg/util/group_role_management_policy_update"
	"github.com/gofrontier-com/sheriff/pkg/util/groups_config"
	"github.com/gofrontier-com/sheriff/pkg/util/user"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// var (
// 	requiredActionsToApply = []string{
// 		"*",
// 		"Microsoft.Authorization/*",
// 	}
// 	requiredActionsToPlan = []string{
// 		"*",
// 		"*/read",
// 		"Microsoft.Authorization/*",
// 		"Microsoft.Authorization/*/read",
// 	}
// )

var (
	dateFormat = "Mon, 02 Jan 2006 15:04:05 MST"
)

//go:embed default_role_management_policy.json
var defaultRoleManagementPolicyPropertiesData string

type GroupsResponse struct {
	Value []GroupResponse `json:"value"`
}

type GroupResponse struct {
	Id string `json:"id"`
}

func ApplyGroups(configDir string, planOnly bool) error {
	var warnings []string

	output.PrintlnInfo("Initialising...")

	output.PrintlnInfo("- Loading and validating config")

	config, err := groups_config.Load(configDir)
	if err != nil {
		if _, ok := err.(*core.ConfigurationEmptyError); ok {
			warnings = append(warnings, "Configuration is empty, is the config path correct?")
		} else {
			return err
		}
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	output.PrintlnfInfo("- Authenticating to Microsoft Graph API")

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	graphServiceClient, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Checking for necessary permissions\n")

	// var requiredActions []string
	// if planOnly {
	// 	requiredActions = requiredActionsToPlan
	// } else {
	// 	requiredActions = requiredActionsToApply
	// }
	err = checkPermissions(graphServiceClient, credential)
	if err != nil {
		return err
	}

	if len(warnings) > 0 {
		output.PrintlnWarn("!!! One or more warnings were generated !!!")

		for _, w := range warnings {
			output.PrintlnfWarn("- %s", w)
		}

		output.Printf("\n")
	}

	output.PrintlnInfo("Sheriff is ready to go!\n")

	output.PrintlnfInfo("Generating plan for...")

	output.PrintlnInfo("- Active assignments")

	// client := &http.Client{}

	// request, err := http.NewRequest("GET", "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups/resources", nil)
	// if err != nil {
	// 	return err
	// }

	// request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// response, err := client.Do(request)
	// if err != nil {
	// 	return err
	// }
	// defer response.Body.Close()

	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	return err
	// }

	// var groupsResponse GroupsResponse
	// if err := json.Unmarshal(body, &groupsResponse); err != nil {
	// 	return err
	// }
	// var groupIds []string
	// linq.From(groupsResponse.Value).SelectT(func(g GroupResponse) string { return g.Id }).ToSlice(&groupIds)
	groupIds := []string{
		"6a7b43df-0876-4cdc-9988-83d9d6a04fb9",
		"d45f33a3-3835-4a58-baa2-9635c004e980",
	}

	groupAssignmentSchedules := config.GetGroupAssignmentSchedules()

	existingGroupGroupAssignmentSchedules, err := group_assignment_schedule.GetGroupAssignmentSchedules(
		graphServiceClient,
		groupIds,
		func(
			graphServiceClient *msgraphsdkgo.GraphServiceClient,
			s models.PrivilegedAccessGroupAssignmentScheduleable,
		) bool {
			principalId := s.GetPrincipalId()
			directoryObject, err := directory_object.GetDirectoryObject(graphServiceClient, *principalId)
			if err != nil {
				panic(err)
			}
			return *directoryObject.GetOdataType() == "#microsoft.graph.group"
		},
	)
	if err != nil {
		return err
	}

	userAssignmentSchedules := config.GetUserAssignmentSchedules()

	existingUserGroupAssignmentSchedules, err := group_assignment_schedule.GetGroupAssignmentSchedules(
		graphServiceClient,
		groupIds,
		func(
			graphServiceClient *msgraphsdkgo.GraphServiceClient,
			s models.PrivilegedAccessGroupAssignmentScheduleable,
		) bool {
			p := s.GetPrincipalId()
			directoryObject, err := directory_object.GetDirectoryObject(graphServiceClient, *p)
			if err != nil {
				panic(err)
			}
			return *directoryObject.GetOdataType() == "#microsoft.graph.user"
		},
	)
	if err != nil {
		return err
	}

	groupAssignmentScheduleCreates, err := group_assignment_schedule_create.GetGroupAssignmentScheduleCreates(
		graphServiceClient,
		groupAssignmentSchedules,
		existingGroupGroupAssignmentSchedules,
		userAssignmentSchedules,
		existingUserGroupAssignmentSchedules,
	)
	if err != nil {
		return err
	}

	groupAssignmentScheduleUpdates, err := group_assignment_schedule_update.GetGroupAssignmentScheduleUpdates(
		graphServiceClient,
		groupAssignmentSchedules,
		existingGroupGroupAssignmentSchedules,
		userAssignmentSchedules,
		existingUserGroupAssignmentSchedules,
	)
	if err != nil {
		return err
	}

	groupAssignmentScheduleDeletes, err := group_assignment_schedule_delete.GetGroupAssignmentScheduleDeletes(
		graphServiceClient,
		groupAssignmentSchedules,
		existingGroupGroupAssignmentSchedules,
		userAssignmentSchedules,
		existingUserGroupAssignmentSchedules,
	)
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Eligible assignments")

	groupEligibilitySchedules := config.GetGroupEligibilitySchedules()

	existingGroupGroupEligibilitySchedules, err := group_eligibility_schedule.GetGroupEligibilitySchedules(
		graphServiceClient,
		groupIds,
		func(
			graphServiceClient *msgraphsdkgo.GraphServiceClient,
			s models.PrivilegedAccessGroupEligibilityScheduleable,
		) bool {
			p := s.GetPrincipalId()
			directoryObject, err := directory_object.GetDirectoryObject(graphServiceClient, *p)
			if err != nil {
				panic(err)
			}
			return *directoryObject.GetOdataType() == "#microsoft.graph.group"
		},
	)
	if err != nil {
		return err
	}

	userEligibilitySchedules := config.GetUserEligibilitySchedules()

	existingUserGroupEligibilitySchedules, err := group_eligibility_schedule.GetGroupEligibilitySchedules(
		graphServiceClient,
		groupIds,
		func(
			graphServiceClient *msgraphsdkgo.GraphServiceClient,
			s models.PrivilegedAccessGroupEligibilityScheduleable,
		) bool {
			p := s.GetPrincipalId()
			directoryObject, err := directory_object.GetDirectoryObject(graphServiceClient, *p)
			if err != nil {
				panic(err)
			}
			return *directoryObject.GetOdataType() == "#microsoft.graph.user"
		},
	)
	if err != nil {
		return err
	}

	groupEligibilityScheduleCreates, err := group_eligibility_schedule_create.GetGroupEligibilityScheduleCreates(
		graphServiceClient,
		groupEligibilitySchedules,
		existingGroupGroupEligibilitySchedules,
		userEligibilitySchedules,
		existingUserGroupEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	groupEligibilityScheduleUpdates, err := group_eligibility_schedule_update.GetGroupEligibilityScheduleUpdates(
		graphServiceClient,
		groupEligibilitySchedules,
		existingGroupGroupEligibilitySchedules,
		userEligibilitySchedules,
		existingUserGroupEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	groupEligibilityScheduleDeletes, err := group_eligibility_schedule_delete.GetGroupEligibilityScheduleDeletes(
		graphServiceClient,
		groupEligibilitySchedules,
		existingGroupGroupEligibilitySchedules,
		userEligibilitySchedules,
		existingUserGroupEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	output.PrintlnInfo("- Role management policies\n")

	roleManagementPolicyUpdates, err := group_role_management_policy_update.GetGroupRoleManagementPolicyUpdates(
		graphServiceClient,
		defaultRoleManagementPolicyPropertiesData,
		config,
	)
	if err != nil {
		return err
	}

	if planOnly {
		output.PrintlnInfo("Sheriff would perform the following actions:\n")
	} else {
		output.PrintlnInfo("Sheriff will perform the following actions:\n")
	}

	printPlan(
		groupAssignmentScheduleCreates,
		groupAssignmentScheduleUpdates,
		groupAssignmentScheduleDeletes,
		groupEligibilityScheduleCreates,
		groupEligibilityScheduleUpdates,
		groupEligibilityScheduleDeletes,
		roleManagementPolicyUpdates,
	)

	if planOnly {
		return nil
	}

	if len(groupAssignmentScheduleCreates)+len(groupAssignmentScheduleUpdates)+len(groupAssignmentScheduleDeletes)+len(groupEligibilityScheduleCreates)+len(groupEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates)+len(groupEligibilityScheduleDeletes) == 0 {
		output.PrintlnInfo("\nNothing to do!")
		return nil
	}

	output.PrintlnInfo("\nApplying plan...\n")

	// for _, u := range roleManagementPolicyUpdates {
	// 	output.PrintlnfInfo(
	// 		"Updating role management policy for role \"%s\" at scope \"%s\"",
	// 		u.RoleName,
	// 		u.Scope,
	// 	)

	// 	_, err = roleManagementPoliciesClient.Update(
	// 		context.Background(),
	// 		u.Scope,
	// 		*u.RoleManagementPolicy.Name,
	// 		*u.RoleManagementPolicy,
	// 		nil,
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for _, c := range groupAssignmentScheduleCreates {
		output.PrintlnfInfo(
			"Creating active assignment for %s \"%s\" with role \"%s\" in managed group \"%s\"",
			c.PrincipalType,
			c.PrincipalName,
			c.RoleName,
			c.ManagedGroupName,
		)

		_, err := graphServiceClient.IdentityGovernance().
			PrivilegedAccess().Group().AssignmentScheduleRequests().
			Post(context.Background(), c.GroupAssignmentScheduleRequest, nil)
		if err != nil {
			return err
		}
	}

	for _, u := range groupAssignmentScheduleUpdates {
		output.PrintlnfInfo(
			"Updating active assignment for %s \"%s\" with role \"%s\" in managed group \"%s\"",
			u.PrincipalType,
			u.PrincipalName,
			u.RoleName,
			u.ManagedGroupName,
		)

		_, err := graphServiceClient.IdentityGovernance().
			PrivilegedAccess().Group().AssignmentScheduleRequests().
			Post(context.Background(), u.GroupAssignmentScheduleRequest, nil)
		if err != nil {
			return err
		}
	}

	for _, d := range groupAssignmentScheduleDeletes {
		output.PrintlnfInfo(
			"Deleting active assignment for %s \"%s\" with role \"%s\" in managed group \"%s\"",
			d.PrincipalType,
			d.PrincipalName,
			d.RoleName,
			d.ManagedGroupName,
		)

		if d.Cancel {
			// TODO: Cancellation request.
		} else {
			_, err := graphServiceClient.IdentityGovernance().
				PrivilegedAccess().Group().AssignmentScheduleRequests().
				Post(context.Background(), d.GroupAssignmentScheduleRequest, nil)
			if err != nil {
				return err
			}
		}
	}

	for _, c := range groupEligibilityScheduleCreates {
		output.PrintlnfInfo(
			"Creating eligible assignment for %s \"%s\" with role \"%s\" in managed group \"%s\"",
			c.PrincipalType,
			c.PrincipalName,
			c.RoleName,
			c.ManagedGroupName,
		)

		_, err := graphServiceClient.IdentityGovernance().
			PrivilegedAccess().Group().EligibilityScheduleRequests().
			Post(context.Background(), c.GroupEligibilityScheduleRequest, nil)
		if err != nil {
			return err
		}
	}

	for _, u := range groupEligibilityScheduleUpdates {
		output.PrintlnfInfo(
			"Updating eligible assignment for %s \"%s\" with role \"%s\" in managed group \"%s\"",
			u.PrincipalType,
			u.PrincipalName,
			u.RoleName,
			u.ManagedGroupName,
		)

		_, err := graphServiceClient.IdentityGovernance().
			PrivilegedAccess().Group().EligibilityScheduleRequests().
			Post(context.Background(), u.GroupEligibilityScheduleRequest, nil)
		if err != nil {
			return err
		}
	}

	for _, d := range groupEligibilityScheduleDeletes {
		output.PrintlnfInfo(
			"Deleting eligible assignment for %s \"%s\" with role \"%s\" in managed group \"%s\"",
			d.PrincipalType,
			d.PrincipalName,
			d.RoleName,
			d.ManagedGroupName,
		)

		if d.Cancel {
			// TODO: Cancellation request.
		} else {
			_, err := graphServiceClient.IdentityGovernance().
				PrivilegedAccess().Group().EligibilityScheduleRequests().
				Post(context.Background(), d.GroupEligibilityScheduleRequest, nil)
			if err != nil {
				return err
			}
		}
	}

	// output.PrintlnfInfo("\nApply complete: %d added, %d changed, %d deleted", len(groupAssignmentScheduleCreates)+len(groupEligibilityScheduleCreates), len(groupAssignmentScheduleUpdates)+len(groupEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates), len(groupAssignmentScheduleDeletes)+len(groupEligibilityScheduleDeletes))

	return nil
}

func checkPermissions(
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	credential *azidentity.DefaultAzureCredential,
) error {
	errors := []string{}

	_, err := user.GetUserByUpn(graphServiceClient, "foo@bar.com")
	if err != nil {
		message := err.Error()
		if message == "Insufficient privileges to complete the operation." {
			errors = append(errors, "at least one of the following microsoft graph permissions are required: User.ReadBasic.All, User.Read.All, Directory.Read.All")
		}
	}

	_, err = group.GetGroupByName(graphServiceClient, "foo bar")
	if err != nil {
		message := err.Error()
		if message == "Insufficient privileges to complete the operation." {
			errors = append(errors, "at least one of the following microsoft graph permissions are required: GroupMember.Read.All, Group.Read.All, Directory.Read.All")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("authenticated principal failed the permissions check with the following error(s):\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

func printPlan(
	groupAssignmentScheduleCreates []*core.GroupAssignmentScheduleCreate,
	groupAssignmentScheduleUpdates []*core.GroupAssignmentScheduleUpdate,
	groupAssignmentScheduleDeletes []*core.GroupAssignmentScheduleDelete,
	groupEligibilityScheduleCreates []*core.GroupEligibilityScheduleCreate,
	groupEligibilityScheduleUpdates []*core.GroupEligibilityScheduleUpdate,
	groupEligibilityScheduleDeletes []*core.GroupEligibilityScheduleDelete,
	roleManagementPolicyUpdates []*core.GroupRoleManagementPolicyUpdate,
) {
	builder := &strings.Builder{}

	if len(groupAssignmentScheduleCreates)+len(groupAssignmentScheduleUpdates)+len(groupAssignmentScheduleDeletes)+len(groupEligibilityScheduleCreates)+len(groupEligibilityScheduleUpdates)+len(groupEligibilityScheduleDeletes)+len(roleManagementPolicyUpdates) == 0 {
		builder.WriteString("(none)\n\n")
	} else {

		if len(groupAssignmentScheduleCreates) > 0 {
			builder.WriteString("  # Create active assignments:\n\n")
			for _, c := range groupAssignmentScheduleCreates {
				builder.WriteString(fmt.Sprintf("    + %s:         %s\n", c.PrincipalType, c.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:          %s\n", c.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n", c.ManagedGroupName))
				builder.WriteString(fmt.Sprintf("      Start:         %s\n", c.StartDateTime.Format(dateFormat)))
				if c.EndDateTime != nil {
					builder.WriteString(fmt.Sprintf("      End:           %s\n", c.EndDateTime.Format(dateFormat)))
				}
				builder.WriteString("\n")
			}
		}

		if len(groupEligibilityScheduleCreates) > 0 {
			builder.WriteString("  # Create eligible assignments:\n\n")
			for _, c := range groupEligibilityScheduleCreates {
				builder.WriteString(fmt.Sprintf("    + %s:         %s\n", c.PrincipalType, c.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:          %s\n", c.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n", c.ManagedGroupName))
				builder.WriteString(fmt.Sprintf("      Start:         %s\n", c.StartDateTime.Format(dateFormat)))
				if c.EndDateTime != nil {
					builder.WriteString(fmt.Sprintf("      End:           %s\n", c.EndDateTime.Format(dateFormat)))
				}
				builder.WriteString("\n")
			}
		}

		if len(groupAssignmentScheduleUpdates) > 0 {
			builder.WriteString("  # Update active assignments:\n\n")
			for _, u := range groupAssignmentScheduleUpdates {
				builder.WriteString(fmt.Sprintf("    ~ %s:         %s\n", u.PrincipalType, u.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:          %s\n", u.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n", u.ManagedGroupName))
				builder.WriteString(fmt.Sprintf("      Start:         %s\n", u.StartDateTime.Format(dateFormat)))
				if u.EndDateTime != nil {
					builder.WriteString(fmt.Sprintf("      End:           %s\n", u.EndDateTime.Format(dateFormat)))
				}
				builder.WriteString("\n")
			}
		}

		if len(groupEligibilityScheduleUpdates) > 0 {
			builder.WriteString("  # Update eligible assignments:\n\n")
			for _, u := range groupEligibilityScheduleUpdates {
				builder.WriteString(fmt.Sprintf("    ~ %s:         %s\n", u.PrincipalType, u.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:          %s\n", u.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n", u.ManagedGroupName))
				builder.WriteString(fmt.Sprintf("      Start:         %s\n", u.StartDateTime.Format(dateFormat)))
				if u.EndDateTime != nil {
					builder.WriteString(fmt.Sprintf("      End:           %s\n", u.EndDateTime.Format(dateFormat)))
				}
				builder.WriteString("\n")
			}
		}

		if len(roleManagementPolicyUpdates) > 0 {
			builder.WriteString("  # Update role management policies:\n\n")
			for _, u := range roleManagementPolicyUpdates {
				builder.WriteString(fmt.Sprintf("    ~ Role:          %s\n", u.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n\n", u.ManagedGroupName))
			}
		}

		if len(groupAssignmentScheduleDeletes) > 0 {
			builder.WriteString("  # Delete active assignments:\n\n")
			for _, d := range groupAssignmentScheduleDeletes {
				builder.WriteString(fmt.Sprintf("    - %s:         %s\n", d.PrincipalType, d.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:          %s\n", d.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n", d.ManagedGroupName))
				builder.WriteString(fmt.Sprintf("      Start:         %s\n", d.StartDateTime.Format(dateFormat)))
				if d.EndDateTime != nil {
					builder.WriteString(fmt.Sprintf("      End:           %s\n", d.EndDateTime.Format(dateFormat)))
				}
				builder.WriteString("\n")
			}
		}

		if len(groupEligibilityScheduleDeletes) > 0 {
			builder.WriteString("  # Delete eligible assignments:\n\n")
			for _, d := range groupEligibilityScheduleDeletes {
				builder.WriteString(fmt.Sprintf("    - %s:         %s\n", d.PrincipalType, d.PrincipalName))
				builder.WriteString(fmt.Sprintf("      Role:          %s\n", d.RoleName))
				builder.WriteString(fmt.Sprintf("      Managed group: %s\n", d.ManagedGroupName))
				builder.WriteString(fmt.Sprintf("      Start:         %s\n", d.StartDateTime.Format(dateFormat)))
				if d.EndDateTime != nil {
					builder.WriteString(fmt.Sprintf("      End:           %s\n", d.EndDateTime.Format(dateFormat)))
				}
				builder.WriteString("\n")
			}
		}
	}

	builder.WriteString(fmt.Sprintf("Plan: %d to add, %d to change, %d to delete.", len(groupAssignmentScheduleCreates)+len(groupEligibilityScheduleCreates), len(groupAssignmentScheduleUpdates)+len(groupEligibilityScheduleUpdates)+len(roleManagementPolicyUpdates), len(groupAssignmentScheduleDeletes)+len(groupEligibilityScheduleDeletes)))

	output.PrintlnInfo(builder.String())
}
