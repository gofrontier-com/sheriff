package apply

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/frontierdigital/sheriff/pkg/core"
	"github.com/frontierdigital/sheriff/pkg/util/azure_rbac_config"
	"github.com/frontierdigital/sheriff/pkg/util/filter"
	"github.com/frontierdigital/sheriff/pkg/util/group"
	"github.com/frontierdigital/sheriff/pkg/util/role_assignment"
	"github.com/frontierdigital/sheriff/pkg/util/role_definition"
	"github.com/frontierdigital/sheriff/pkg/util/role_eligibility_schedule"
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

func ApplyAzureRbac(configDir string, subscriptionId string, dryRun bool) error {
	PrintHeader(configDir, subscriptionId, dryRun)

	output.PrintlnfInfo("Loading and validating Azure Rbac config from %s", configDir)

	config, err := azure_rbac_config.Load(configDir)
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

	clientFactory, err := armauthorization.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return err
	}

	// roleAssignmentSchedulesClient := clientFactory.NewRoleAssignmentSchedulesClient()

	graphServiceClient, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return err
	}

	CheckPermissions(clientFactory, graphServiceClient, subscriptionId)

	output.PrintlnfInfo("Comparing config against subscription with Id \"%s\"", subscriptionId)

	activeAssignments := config.GetActiveAssignments(subscriptionId)

	existingRoleAssignments, err := role_assignment.GetRoleAssignments(
		clientFactory,
		cache,
		subscriptionId,
		func(a *armauthorization.RoleAssignment) bool {
			return *a.Properties.PrincipalType == armauthorization.PrincipalTypeGroup
		},
	)
	if err != nil {
		return err
	}

	activeAssignmentsToCreate, err := getActiveAssignmentsToCreate(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		activeAssignments,
		existingRoleAssignments,
	)
	if err != nil {
		return err
	}

	activeAssignmentsToIgnore, err := getActiveAssignmentsToIgnore(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		activeAssignments,
		existingRoleAssignments,
	)
	if err != nil {
		return err
	}

	roleAssignmentsToDelete, err := getRoleAssignmentsToDelete(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		activeAssignments,
		existingRoleAssignments,
	)
	if err != nil {
		return err
	}

	eligibleAssignments := config.GetEligibleAssignments(subscriptionId)
	_ = eligibleAssignments

	existingRoleEligibilitySchedules, err := role_eligibility_schedule.GetRoleEligibilitySchedules(
		clientFactory,
		cache,
		subscriptionId,
		func(s *armauthorization.RoleEligibilitySchedule) bool {
			return *s.Properties.PrincipalType == armauthorization.PrincipalTypeGroup
		},
	)
	if err != nil {
		return err
	}

	eligibleAssignmentsToCreate, err := getEligibleAssignmentsToCreate(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		eligibleAssignments,
		existingRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	roleEligibilitySchedulesToDelete, err := getRoleEligibilitySchedulesToDelete(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		eligibleAssignments,
		existingRoleEligibilitySchedules,
	)
	if err != nil {
		return err
	}

	output.PrintlnInfo(strings.Repeat("-", 78))
	output.PrintlnfInfo("Active assignments to create: %d", len(activeAssignmentsToCreate))
	output.PrintlnfInfo("Active assignments to delete: %d", len(roleAssignmentsToDelete))
	output.PrintlnfInfo("Eligible assignments to create: %d", len(eligibleAssignmentsToCreate))
	output.PrintlnfInfo("Eligible assignments to update: %d", "N/A")
	output.PrintlnfInfo("Eligible assignments to delete: %d", len(roleEligibilitySchedulesToDelete))
	output.PrintlnfInfo("\nActive assignments unchanged: %d", len(activeAssignmentsToIgnore))
	output.PrintlnfInfo("Eligible assignments unchanged: %d", "N/A")
	output.PrintlnInfo(strings.Repeat("-", 78))

	// PrintPlan(len(activeAssignmentsToCreate), len(roleAssignmentsToDelete), len(activeAssignmentsToIgnore))

	if dryRun {
		return nil
	}

	err = createActiveAssignments(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		activeAssignmentsToCreate,
	)
	if err != nil {
		return err
	}

	err = deleteRoleAssignments(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		roleAssignmentsToDelete,
	)
	if err != nil {
		return err
	}

	err = createEligibleAssignments(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		eligibleAssignmentsToCreate,
	)
	if err != nil {
		return err
	}

	err = deleteRoleEligibilitySchedules(
		clientFactory,
		graphServiceClient,
		subscriptionId,
		roleEligibilitySchedulesToDelete,
	)
	if err != nil {
		return err
	}

	return nil
}

func createActiveAssignments(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	activeAssignments []*core.ActiveAssignment,
) error {
	roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()

	for _, a := range activeAssignments {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, cache, subscriptionId, a.RoleName)
		if err != nil {
			return err
		}

		group, err := group.GetGroupByName(graphServiceClient, cache, a.GroupName)
		if err != nil {
			return err
		}

		output.PrintlnfInfo(
			"Creating active assignment for group \"%s\" with role \"%s\" at scope \"%s\"",
			a.GroupName,
			a.RoleName,
			a.Scope,
		)

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
			return err
		}
	}

	return nil
}

func deleteRoleAssignments(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	roleAssignments []*armauthorization.RoleAssignment,
) error {
	roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()

	for _, r := range roleAssignments {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			cache,
			subscriptionId,
			*r.Properties.RoleDefinitionID,
		)
		if err != nil {
			return err
		}

		group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
		if err != nil {
			return err
		}

		output.PrintlnfInfo(
			"Deleting active assignment for group \"%s\" with role \"%s\" at scope \"%s\"",
			*group.GetDisplayName(),
			*roleDefinition.Properties.RoleName,
			*r.Properties.Scope,
		)

		_, err = roleAssignmentsClient.DeleteByID(context.Background(), *r.ID, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func createEligibleAssignments(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	eligibleAssignments []*core.EligibleAssignment,
) error {
	roleEligibilityScheduleRequestsClient := clientFactory.NewRoleEligibilityScheduleRequestsClient()

	for _, e := range eligibleAssignments {
		roleDefinition, err := role_definition.GetRoleDefinitionByName(clientFactory, cache, subscriptionId, e.RoleName)
		if err != nil {
			return err
		}

		group, err := group.GetGroupByName(graphServiceClient, cache, e.GroupName)
		if err != nil {
			return err
		}

		output.PrintlnfInfo(
			"Creating eligible assignment for group \"%s\" with role \"%s\" at scope \"%s\"",
			e.GroupName,
			e.RoleName,
			e.Scope,
		)

		id := uuid.New()
		roleEligibilityScheduleRequest := &armauthorization.RoleEligibilityScheduleRequest{
			Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
				PrincipalID:      group.GetId(),
				RequestType:      to.Ptr(armauthorization.RequestTypeAdminAssign),
				RoleDefinitionID: roleDefinition.ID,
				ScheduleInfo:     getScheduleInfo(e.StartDateTime, e.EndDateTime),
			},
		}
		_, err = roleEligibilityScheduleRequestsClient.Create(
			context.Background(),
			fmt.Sprintf("/providers/Microsoft.Subscription%s", e.Scope),
			id.String(),
			*roleEligibilityScheduleRequest,
			nil,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteRoleEligibilitySchedules(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	roleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) error {
	roleEligibilityScheduleRequestsClient := clientFactory.NewRoleEligibilityScheduleRequestsClient()

	for _, r := range roleEligibilitySchedules {
		roleDefinition, err := role_definition.GetRoleDefinitionById(
			clientFactory,
			cache,
			subscriptionId,
			*r.Properties.RoleDefinitionID,
		)
		if err != nil {
			return err
		}

		group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
		if err != nil {
			return err
		}

		id := uuid.New()
		output.PrintlnfInfo(
			"Deleting eligible assignment for group \"%s\" with role \"%s\" at scope \"%s\"",
			*group.GetDisplayName(),
			*roleDefinition.Properties.RoleName,
			*r.Properties.Scope,
		)

		if *r.Properties.Status == armauthorization.StatusProvisioned {
			roleEligibilityScheduleRequest := &armauthorization.RoleEligibilityScheduleRequest{
				Properties: &armauthorization.RoleEligibilityScheduleRequestProperties{
					PrincipalID:                     r.Properties.PrincipalID,
					RequestType:                     to.Ptr(armauthorization.RequestTypeAdminRemove),
					RoleDefinitionID:                r.Properties.RoleDefinitionID,
					TargetRoleEligibilityScheduleID: r.ID, // TODO: Should this be TargetRoleEligibilityScheduleInstanceID?
				},
			}
			_, err = roleEligibilityScheduleRequestsClient.Create(
				context.Background(),
				fmt.Sprintf("/providers/Microsoft.Subscription%s", *r.Properties.Scope),
				id.String(),
				*roleEligibilityScheduleRequest,
				nil,
			)
			if err != nil {
				return err
			}
		} else {
			_, err = roleEligibilityScheduleRequestsClient.Cancel(
				context.Background(),
				fmt.Sprintf("/providers/Microsoft.Subscription%s", *r.Properties.Scope),
				*r.Name,
				nil,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getActiveAssignmentsToCreate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	activeAssignments []*core.ActiveAssignment,
	existingRoleAssignments []*armauthorization.RoleAssignment,
) (filtered []*core.ActiveAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(activeAssignments, func(a *core.ActiveAssignment) bool {
		idx := slices.IndexFunc(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				subscriptionId,
				*r.Properties.RoleDefinitionID,
			)
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

func getActiveAssignmentsToIgnore(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	activeAssignments []*core.ActiveAssignment,
	existingRoleAssignments []*armauthorization.RoleAssignment,
) (filtered []*core.ActiveAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(activeAssignments, func(a *core.ActiveAssignment) bool {
		idx := slices.IndexFunc(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				subscriptionId,
				*r.Properties.RoleDefinitionID,
			)
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

func getEligibleAssignmentsToCreate(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	eligibleAssignments []*core.EligibleAssignment,
	existingRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) (filtered []*core.EligibleAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(eligibleAssignments, func(e *core.EligibleAssignment) bool {
		idx := slices.IndexFunc(existingRoleEligibilitySchedules, func(s *armauthorization.RoleEligibilitySchedule) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				subscriptionId,
				*s.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, cache, *s.Properties.PrincipalID) // TODO: Get group Id by display name instead?
			if err != nil {
				panic(err)
			}

			return e.Scope == *s.Properties.Scope &&
				e.RoleName == *roleDefinition.Properties.RoleName &&
				e.GroupName == *group.GetDisplayName()
		})

		return idx == -1
	})

	return
}

func getRoleAssignmentsToDelete(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	activeAssignments []*core.ActiveAssignment,
	existingRoleAssignments []*armauthorization.RoleAssignment,
) (filtered []*armauthorization.RoleAssignment, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(existingRoleAssignments, func(r *armauthorization.RoleAssignment) bool {
		idx := slices.IndexFunc(activeAssignments, func(a *core.ActiveAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				subscriptionId,
				*r.Properties.RoleDefinitionID,
			)
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

func getRoleEligibilitySchedulesToDelete(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
	eligibleAssignments []*core.EligibleAssignment,
	existingRoleEligibilitySchedules []*armauthorization.RoleEligibilitySchedule,
) (filtered []*armauthorization.RoleEligibilitySchedule, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	filtered = filter.Filter(existingRoleEligibilitySchedules, func(r *armauthorization.RoleEligibilitySchedule) bool {
		idx := slices.IndexFunc(eligibleAssignments, func(e *core.EligibleAssignment) bool {
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				subscriptionId,
				*r.Properties.RoleDefinitionID,
			)
			if err != nil {
				panic(err)
			}

			group, err := group.GetGroupById(graphServiceClient, cache, *r.Properties.PrincipalID)
			if err != nil {
				panic(err)
			}

			return *r.Properties.Scope == e.Scope &&
				*roleDefinition.Properties.RoleName == e.RoleName &&
				*group.GetDisplayName() == e.GroupName
		})

		return idx == -1
	})

	return
}

func getScheduleInfo(
	startDateTime *time.Time,
	endDateTime *time.Time,
) *armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo {
	var expiration armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration

	if endDateTime == nil {
		expiration = armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration{
			Type: to.Ptr(armauthorization.TypeNoExpiration),
		}
	} else {
		expiration = armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration{
			EndDateTime: endDateTime,
			Type:        to.Ptr(armauthorization.TypeAfterDateTime),
		}
	}

	if startDateTime == nil {
		return &armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo{
			Expiration:    &expiration,
			StartDateTime: to.Ptr(time.Now()),
		}
	} else {
		return &armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo{
			Expiration:    &expiration,
			StartDateTime: startDateTime,
		}
	}
}

func CheckPermissions(
	clientFactory *armauthorization.ClientFactory,
	graphServiceClient *msgraphsdkgo.GraphServiceClient,
	subscriptionId string,
) error {
	me, err := graphServiceClient.Me().Get(context.Background(), nil)
	if err != nil {
		return err
	}

	roleAssignmentsClient := clientFactory.NewRoleAssignmentsClient()

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
			roleDefinition, err := role_definition.GetRoleDefinitionById(
				clientFactory,
				cache,
				subscriptionId,
				*r.Properties.RoleDefinitionID,
			)
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
