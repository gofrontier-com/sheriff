package core

type activeAssignments struct {
	ResourceGroups map[string][]*ActiveAssignment `yaml:"resourceGroups" validate:"dive"`
	Resources      map[string][]*ActiveAssignment `yaml:"resources" validate:"dive"`
	Subscription   []*ActiveAssignment            `yaml:"subscription" validate:"dive"`
}

type AppConfig struct {
}

type ActiveAssignment struct {
	GroupName string
	RoleName  string `yaml:"roleName" validate:"required"`
	Scope     string
}

type Config struct {
	Groups   []*Group  `validate:"dive"`
	Policies []*Policy `validate:"dive"`
}

type EligibleAssignment struct {
	EndDateTime              string `yaml:"endDateTime" validate:"datetime"`
	GroupName                string
	RoleManagementPolicyName string `yaml:"roleManagementPolicyName"`
	RoleName                 string `yaml:"roleName" validate:"required"`
	Scope                    string
}

type eligibleAssignments struct {
	ResourceGroups map[string][]*EligibleAssignment `yaml:"resourceGroups" validate:"dive"`
	Resources      map[string][]*EligibleAssignment `yaml:"resources" validate:"dive"`
	Subscription   []*EligibleAssignment            `yaml:"subscription" validate:"dive"`
}

type Group struct {
	Active   *activeAssignments   `yaml:"active"`
	Eligible *eligibleAssignments `yaml:"eligible"`
	Name     string
}

type Policy struct {
}
