package apply

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/frontierdigital/sheriff/pkg/core"
	"github.com/frontierdigital/sheriff/pkg/util/config"
	"github.com/frontierdigital/sheriff/pkg/util/filter"
	"github.com/frontierdigital/sheriff/pkg/util/group"
	"github.com/frontierdigital/sheriff/pkg/util/role_assignment"
	"github.com/frontierdigital/sheriff/pkg/util/role_assignment_schedule"
	"github.com/frontierdigital/sheriff/pkg/util/role_definition"
	"github.com/frontierdigital/utils/output"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/exp/slices"
)

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func Apply(configDir string, subscriptionId string, dryRun bool) error {
	PrintHeader(configDir, subscriptionId, dryRun)

	output.PrintlnfInfo("Loading and validating config from %s", configDir)

	config, err := config.LoadConfig(configDir)
	if err != nil {
		return err
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	output.PrintlnfInfo("Authenticating to the Azure Management API and checking for necessary permissions")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	roleAssignmentsClient, err := armauthorization.NewRoleAssignmentsClient(subscriptionId, cred, nil)
	if err != nil {
		return err
	}

	roleAssignmentSchedulesClient, err := armauthorization.NewRoleAssignmentSchedulesClient(cred, nil)
	if err != nil {
		return err
	}

	roleDefinitionsClient, err := armauthorization.NewRoleDefinitionsClient(cred, nil)
	if err != nil {
		return err
	}

	graphServiceClient, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return err
	}

	CheckPermissions(roleAssignmentsClient, roleDefinitionsClient, graphServiceClient, subscriptionId)

	output.PrintlnfInfo("Comparing config against subscription with Id \"%s\"", subscriptionId)

	activeAssignments := config.GetActiveAssignments(subscriptionId)

	existingRoleAssignments, err := role_assignment.GetRoleAssignments(roleAssignmentsClient, cache, subscriptionId, func(a *armauthorization.RoleAssignment) bool {
		return *a.Properties.PrincipalType == armauthorization.PrincipalTypeGroup
	})
	if err != nil {
		return err
	}

	activeAssignmentsToCreate, err := getActiveAssignmentsToCreate(roleDefinitionsClient, graphServiceClient, subscriptionId, activeAssignments, existingRoleAssignments)
	if err != nil {
		return err
	}

	activeAssignmentsToIgnore, err := getActiveAssignmentsToIgnore(roleDefinitionsClient, graphServiceClient, subscriptionId, activeAssignments, existingRoleAssignments)
	if err != nil {
		return err
	}

	roleAssignmentsToDelete, err := getRoleAssignmentsToDelete(roleDefinitionsClient, graphServiceClient, subscriptionId, activeAssignments, existingRoleAssignments)
	if err != nil {
		return err
	}

	eligibleAssignments := config.GetEligibleAssignments(subscriptionId)
	_ = eligibleAssignments

	existingRoleAssignmentSchedules, err := role_assignment_schedule.GetRoleAssignmentSchedules(roleAssignmentSchedulesClient, cache, subscriptionId, func(s *armauthorization.RoleAssignmentSchedule) bool {
		return true
	})
	if err != nil {
		return err
	}
	_ = existingRoleAssignmentSchedules

	PrintPlan(len(activeAssignmentsToCreate), len(roleAssignmentsToDelete), len(activeAssignmentsToIgnore))

	if dryRun {
		return nil
	}

	for _, a := range activeAssignmentsToCreate {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(roleDefinitionsClient, cache, subscriptionId, a.RoleName)
		if err != nil {
			panic(err)
		}

		group, err := group.GetGroupByName(graphServiceClient, cache, a.GroupName)
		if err != nil {
			panic(err)
		}

		output.PrintlnfInfo("Creating active assignment for group \"%s\" with role \"%s\" at scope \"%s\"", a.GroupName, a.RoleName, a.Scope)

		id := uuid.New()
		roleAssignmentCreateParameters := &armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				Description:      to.Ptr("Managed by Sheriff"),
				PrincipalID:      group.GetId(),
				PrincipalType:    to.Ptr(armauthorization.PrincipalTypeGroup),
				RoleDefinitionID: roleDefinition.ID,
			},
		}
		_, err = roleAssignmentsClient.Create(context.Background(), a.Scope, id.String(), *roleAssignmentCreateParameters, nil)
		if err != nil {
			panic(err)
		}
	}

	for _, r := range roleAssignmentsToDelete {
		roleDefinition, err := role_definition.GetRoleDefinitionById(roleDefinitionsClient, cache, subscriptionId, *r.Properties.RoleDefinitionID)
		if err != nil {
			panic(err)
		}

		group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
		if err != nil {
			panic(err)
		}

		output.PrintlnfInfo("Deleting active assignment for group \"%s\" with role \"%s\" at scope \"%s\"", *group.GetDisplayName(), *roleDefinition.Properties.RoleName, *r.Properties.Scope)

		_, err = roleAssignmentsClient.DeleteByID(context.Background(), *r.ID, nil)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func getActiveAssignmentsToCreate(roleDefinitionsClient *armauthorization.RoleDefinitionsClient, graphServiceClient *msgraphsdkgo.GraphServiceClient, subscriptionId string, activeAssignments []*core.ActiveAssignment, existingRoleAssignments []*armauthorization.RoleAssignment) (filtered []*core.ActiveAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(activeAssignments, func(a *core.ActiveAssignment) bool {
		idx := slices.IndexFunc(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(roleDefinitionsClient, cache, subscriptionId, *r.Properties.RoleDefinitionID)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
			if err != nil {
				panic(err)
			}

			return a.Scope == *r.Properties.Scope &&
				a.RoleName == *roleDefinition.Properties.RoleName &&
				a.GroupName == *group.GetDisplayName()
		})

		return idx == -1
	})

	return
}

func getActiveAssignmentsToIgnore(roleDefinitionsClient *armauthorization.RoleDefinitionsClient, graphServiceClient *msgraphsdkgo.GraphServiceClient, subscriptionId string, activeAssignments []*core.ActiveAssignment, existingRoleAssignments []*armauthorization.RoleAssignment) (filtered []*core.ActiveAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(activeAssignments, func(a *core.ActiveAssignment) bool {
		idx := slices.IndexFunc(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(roleDefinitionsClient, cache, subscriptionId, *r.Properties.RoleDefinitionID)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
			if err != nil {
				panic(err)
			}

			return a.Scope == *r.Properties.Scope &&
				a.RoleName == *roleDefinition.Properties.RoleName &&
				a.GroupName == *group.GetDisplayName()
		})

		return idx != -1
	})

	return
}

func getRoleAssignmentsToDelete(roleDefinitionsClient *armauthorization.RoleDefinitionsClient, graphServiceClient *msgraphsdkgo.GraphServiceClient, subscriptionId string, activeAssignments []*core.ActiveAssignment, existingRoleAssignments []*armauthorization.RoleAssignment) (filtered []*armauthorization.RoleAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
		idx := slices.IndexFunc(activeAssignments, func(a *core.ActiveAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(roleDefinitionsClient, cache, subscriptionId, *r.Properties.RoleDefinitionID)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
			if err != nil {
				panic(err)
			}

			return *r.Properties.Scope == a.Scope &&
				*roleDefinition.Properties.RoleName == a.RoleName &&
				*group.GetDisplayName() == a.GroupName
		})

		return idx == -1
	})

	return
}

func CheckPermissions(roleAssignmentsClient *armauthorization.RoleAssignmentsClient, roleDefinitionsClient *armauthorization.RoleDefinitionsClient, graphServiceClient *msgraphsdkgo.GraphServiceClient, subscriptionId string) error {
	me, err := graphServiceClient.Me().Get(context.Background(), nil)
	if err != nil {
		return err
	}

	hasRoleAssignmentsWritePermission := false
	roleAssignmentsClientListForSubscriptionOptions := &armauthorization.RoleAssignmentsClientListForSubscriptionOptions{
		Filter: to.Ptr(fmt.Sprintf("assignedTo('%s')", *me.GetId())),
	}
	pager := roleAssignmentsClient.NewListForSubscriptionPager(roleAssignmentsClientListForSubscriptionOptions)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return err
		}

		for _, r := range page.Value {
			roleDefinition, err := role_definition.GetRoleDefinitionById(roleDefinitionsClient, cache, subscriptionId, *r.Properties.RoleDefinitionID)
			if err != nil {
				return err
			}

			for _, p := range roleDefinition.Properties.Permissions {
				for _, a := range p.Actions {
					if *a == "*" {
						hasRoleAssignmentsWritePermission = true
						break
					}
					if *a == "Microsoft.Authorization/roleAssignments/write" {
						hasRoleAssignmentsWritePermission = true
						break
					}
				}
			}
		}
	}

	if !hasRoleAssignmentsWritePermission {
		return fmt.Errorf("user does not have permission to create role assignments") // TODO: Update
	}

	return nil
}

func PrintHeader(configDir string, subscriptionId string, dryRun bool) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 78)))
	builder.WriteString(fmt.Sprintf("Config directory  | %s\n", configDir))
	builder.WriteString(fmt.Sprintf("Subscription      | %s\n", subscriptionId))
	builder.WriteString(fmt.Sprintf("Mode              | %s\n", "Groups only"))
	builder.WriteString(fmt.Sprintf("Dry-run           | %v\n", dryRun))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 78)))
	output.Println(builder.String())
}

func PrintPlan(createCount int, deleteCount int, ignoreCount int) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("-", 78)))
	builder.WriteString(fmt.Sprintf("|%s|\n", strings.Repeat(" ", 76)))
	builder.WriteString(fmt.Sprintf("|         Create: %d%sDelete: %d%sIgnore: %d          |\n", createCount, strings.Repeat(" ", 15), deleteCount, strings.Repeat(" ", 15), ignoreCount))
	builder.WriteString(fmt.Sprintf("|%s|\n", strings.Repeat(" ", 76)))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 78)))
	output.Println(builder.String())
}
