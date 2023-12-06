package core

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

type activeAssignments struct {
	ResourceGroups map[string][]*ActiveAssignment `yaml:"resourceGroups" validate:"dive"`
	Resources      map[string][]*ActiveAssignment `yaml:"resources" validate:"dive"`
	Subscription   []*ActiveAssignment            `yaml:"subscription" validate:"dive"`
}

type AppConfig struct {
}

type ActiveAssignment struct {
	PrincipalName string
	RoleName      string `yaml:"roleName" validate:"required"`
	Scope         string
}

type AzureRbacConfig struct {
	Groups   []*Principal `validate:"dive"`
	Policies []*Policy    `validate:"dive"`
	Users    []*Principal `validate:"dive"`
}

type EligibleAssignment struct {
	EndDateTime              *time.Time `yaml:"endDateTime"`
	PrincipalName            string
	RoleManagementPolicyName *string `yaml:"roleManagementPolicyName"`
	RoleName                 string  `yaml:"roleName" validate:"required"`
	Scope                    string
	StartDateTime            *time.Time `yaml:"startDateTime"`
}

type eligibleAssignments struct {
	ResourceGroups map[string][]*EligibleAssignment `yaml:"resourceGroups" validate:"dive"`
	Resources      map[string][]*EligibleAssignment `yaml:"resources" validate:"dive"`
	Subscription   []*EligibleAssignment            `yaml:"subscription" validate:"dive"`
}

type Principal struct {
	Active   *activeAssignments   `yaml:"active"`
	Eligible *eligibleAssignments `yaml:"eligible"`
	Name     string
}

type Policy struct {
}

type RoleAssignmentCreate struct {
	PrincipalName                  string
	PrincipalType                  armauthorization.PrincipalType
	RoleAssignmentCreateParameters *armauthorization.RoleAssignmentCreateParameters
	RoleAssignmentName             string
	RoleName                       string
	Scope                          string
}

type RoleAssignmentDelete struct {
	PrincipalName    string
	PrincipalType    armauthorization.PrincipalType
	RoleAssignmentID string
	RoleName         string
	Scope            string
}

type RoleEligibilityScheduleCreate struct {
	PrincipalName                      string
	PrincipalType                      armauthorization.PrincipalType
	RoleEligibilityScheduleRequest     *armauthorization.RoleEligibilityScheduleRequest
	RoleEligibilityScheduleRequestName string
	RoleName                           string
	Scope                              string
}

type RoleEligibilityScheduleDelete struct {
	Cancel bool
	// Id                                 *string
	PrincipalName                      string
	PrincipalType                      armauthorization.PrincipalType
	RoleEligibilityScheduleRequest     *armauthorization.RoleEligibilityScheduleRequest
	RoleEligibilityScheduleRequestName string
	RoleName                           string
	Scope                              string
}
