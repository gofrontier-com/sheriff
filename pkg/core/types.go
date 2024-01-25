package core

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

type AzureRmConfig struct {
	Groups                       []*Principal                   `validate:"dive"` // TODO: Make private
	RoleManagementPolicyRulesets []*RoleManagementPolicyRuleset `validate:"dive"` // TODO: Make private
	Users                        []*Principal                   `validate:"dive"` // TODO: Make private
}

type Principal struct {
	Name           string
	Subscription   *ScopeConfiguration            `yaml:"subscription"`
	ResourceGroups map[string]*ScopeConfiguration `yaml:"resourceGroups"`
	Resources      map[string]*ScopeConfiguration `yaml:"resources"`
}

type ScopeConfiguration struct {
	Active   []*Schedule                    `yaml:"active"`
	Eligible []*Schedule                    `yaml:"eligible"`
	Policy   map[string][]*RulesetReference `yaml:"policy"`
}

type Schedule struct {
	EndDateTime                     *time.Time `yaml:"endDateTime"`
	PrincipalName                   string
	RoleManagementPolicyRulesetName *string `yaml:"roleManagementPolicyRulesetName"`
	RoleName                        string  `yaml:"roleName" validate:"required"`
	Scope                           string
	StartDateTime                   *time.Time `yaml:"startDateTime"`
}

type RulesetReference struct {
	RulesetName string `yaml:"rulesetName"`
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

type RoleManagementPolicyRule struct {
	ID    string      `yaml:"id"`
	Patch interface{} `yaml:"patch"`
}

type RoleManagementPolicyRuleset struct {
	Name  string                      `yaml:"name"`
	Rules []*RoleManagementPolicyRule `yaml:"rules"`
}

type RoleManagementPolicyUpdate struct {
	RoleManagementPolicy *armauthorization.RoleManagementPolicy
	RoleName             string
	Scope                string
}

type ConfigurationEmptyError struct{}
