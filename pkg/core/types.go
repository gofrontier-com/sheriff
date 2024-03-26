package core

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

//region shared

type Schedule struct {
	EndDateTime   *time.Time `yaml:"endDateTime"`
	PrincipalName string
	RoleName      string `yaml:"roleName" validate:"required"`
	Target        string
	StartDateTime *time.Time `yaml:"startDateTime"`
}

type RoleManagementPolicyRule struct {
	ID    string      `yaml:"id" validate:"required"`
	Patch interface{} `yaml:"patch" validate:"required"`
}

type RoleManagementPolicyRuleset struct {
	Name  string
	Rules []*RoleManagementPolicyRule `yaml:"rules"`
}

type RulesetReference struct {
	RulesetName string `yaml:"rulesetName" validate:"required"`
}

type TargetRoleNameCombination struct {
	RoleName string
	Target   string
}

type ConfigurationEmptyError struct{}

//endregion

//region groups

type GroupsConfig struct {
	Groups   []*GroupPrincipal              `validate:"dive"`
	Policies []*GroupPolicy                 `validate:"dive"`
	Rulesets []*RoleManagementPolicyRuleset `validate:"dive"`
	Users    []*GroupPrincipal              `validate:"dive"`
}

type GroupPrincipal struct {
	Name          string
	ManagedGroups map[string]*GroupConfiguration `yaml:"managedGroups"`
}

type GroupConfiguration struct {
	Active   []*Schedule `yaml:"active"`
	Eligible []*Schedule `yaml:"eligible"`
}

type GroupPolicy struct {
	Default       []*RulesetReference `yaml:"default"`
	Name          string
	ManagedGroups map[string][]*RulesetReference `yaml:"managedGroups"`
}

type GroupAssignmentScheduleCreate struct {
	EndDateTime                    *time.Time
	PrincipalName                  string
	PrincipalType                  armauthorization.PrincipalType
	GroupAssignmentScheduleRequest *models.PrivilegedAccessGroupAssignmentScheduleRequest
	ManagedGroupName               string
	RoleName                       string
	StartDateTime                  *time.Time
}

type GroupAssignmentScheduleDelete struct {
	Cancel                         bool
	EndDateTime                    *time.Time
	PrincipalName                  string
	PrincipalType                  armauthorization.PrincipalType
	GroupAssignmentScheduleRequest *models.PrivilegedAccessGroupAssignmentScheduleRequest
	ManagedGroupName               string
	RoleName                       string
	StartDateTime                  *time.Time
}

type GroupAssignmentScheduleUpdate struct {
	EndDateTime                    *time.Time
	PrincipalName                  string
	PrincipalType                  armauthorization.PrincipalType
	GroupAssignmentScheduleRequest *models.PrivilegedAccessGroupAssignmentScheduleRequest
	ManagedGroupName               string
	RoleName                       string
	StartDateTime                  *time.Time
}

type GroupEligibilityScheduleCreate struct {
	EndDateTime                     *time.Time
	PrincipalName                   string
	PrincipalType                   armauthorization.PrincipalType
	GroupEligibilityScheduleRequest *models.PrivilegedAccessGroupEligibilityScheduleRequest
	ManagedGroupName                string
	RoleName                        string
	StartDateTime                   *time.Time
}

type GroupEligibilityScheduleDelete struct {
	Cancel                          bool
	EndDateTime                     *time.Time
	PrincipalName                   string
	PrincipalType                   armauthorization.PrincipalType
	GroupEligibilityScheduleRequest *models.PrivilegedAccessGroupEligibilityScheduleRequest
	ManagedGroupName                string
	RoleName                        string
	StartDateTime                   *time.Time
}

type GroupEligibilityScheduleUpdate struct {
	EndDateTime                     *time.Time
	PrincipalName                   string
	PrincipalType                   armauthorization.PrincipalType
	GroupEligibilityScheduleRequest *models.PrivilegedAccessGroupEligibilityScheduleRequest
	ManagedGroupName                string
	RoleName                        string
	StartDateTime                   *time.Time
}

type GroupRoleManagementPolicyUpdate struct {
	ManagedGroupName     string
	RoleManagementPolicy models.UnifiedRoleManagementPolicy
	RoleName             string
}

//endregion

//region resources

type ResourcesConfig struct {
	Groups   []*ResourcePrincipal           `validate:"dive"`
	Policies []*ResourcePolicy              `validate:"dive"`
	Rulesets []*RoleManagementPolicyRuleset `validate:"dive"`
	Users    []*ResourcePrincipal           `validate:"dive"`
}

type ResourcePrincipal struct {
	Name           string
	Subscription   *ResourceConfiguration            `yaml:"subscription"`
	ResourceGroups map[string]*ResourceConfiguration `yaml:"resourceGroups"`
	Resources      map[string]*ResourceConfiguration `yaml:"resources"`
}

type ResourceConfiguration struct {
	Active   []*Schedule `yaml:"active"`
	Eligible []*Schedule `yaml:"eligible"`
}

type ResourcePolicy struct {
	Default        []*RulesetReference `yaml:"default"`
	Name           string
	Subscription   []*RulesetReference            `yaml:"subscription"`
	ResourceGroups map[string][]*RulesetReference `yaml:"resourceGroups"`
	Resources      map[string][]*RulesetReference `yaml:"resources"`
}

type RoleAssignmentScheduleCreate struct {
	EndDateTime                       *time.Time
	PrincipalName                     string
	PrincipalType                     armauthorization.PrincipalType
	RoleAssignmentScheduleRequest     *armauthorization.RoleAssignmentScheduleRequest
	RoleAssignmentScheduleRequestName string
	RoleName                          string
	Scope                             string
	StartDateTime                     *time.Time
}

type RoleAssignmentScheduleDelete struct {
	Cancel                            bool
	EndDateTime                       *time.Time
	PrincipalName                     string
	PrincipalType                     armauthorization.PrincipalType
	RoleAssignmentScheduleRequest     *armauthorization.RoleAssignmentScheduleRequest
	RoleAssignmentScheduleRequestName string
	RoleName                          string
	Scope                             string
	StartDateTime                     *time.Time
}

type RoleAssignmentScheduleUpdate struct {
	EndDateTime                       *time.Time
	PrincipalName                     string
	PrincipalType                     armauthorization.PrincipalType
	RoleAssignmentScheduleRequest     *armauthorization.RoleAssignmentScheduleRequest
	RoleAssignmentScheduleRequestName string
	RoleName                          string
	Scope                             string
	StartDateTime                     *time.Time
}

type RoleEligibilityScheduleCreate struct {
	EndDateTime                        *time.Time
	PrincipalName                      string
	PrincipalType                      armauthorization.PrincipalType
	RoleEligibilityScheduleRequest     *armauthorization.RoleEligibilityScheduleRequest
	RoleEligibilityScheduleRequestName string
	RoleName                           string
	Scope                              string
	StartDateTime                      *time.Time
}

type RoleEligibilityScheduleDelete struct {
	Cancel                             bool
	EndDateTime                        *time.Time
	PrincipalName                      string
	PrincipalType                      armauthorization.PrincipalType
	RoleEligibilityScheduleRequest     *armauthorization.RoleEligibilityScheduleRequest
	RoleEligibilityScheduleRequestName string
	RoleName                           string
	Scope                              string
	StartDateTime                      *time.Time
}

type RoleEligibilityScheduleUpdate struct {
	EndDateTime                        *time.Time
	PrincipalName                      string
	PrincipalType                      armauthorization.PrincipalType
	RoleEligibilityScheduleRequest     *armauthorization.RoleEligibilityScheduleRequest
	RoleEligibilityScheduleRequestName string
	RoleName                           string
	Scope                              string
	StartDateTime                      *time.Time
}

type ResourceRoleManagementPolicyUpdate struct {
	RoleManagementPolicy *armauthorization.RoleManagementPolicy
	RoleName             string
	Scope                string
}

//endregion
