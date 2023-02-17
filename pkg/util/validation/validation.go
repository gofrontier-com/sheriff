package validation

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/frontierdigital/utils/output"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

var (
	DateTimeRegex = regexp.MustCompile(`^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z$`)
)

func validateDateTime(field validator.FieldLevel) bool {
	return DateTimeRegex.MatchString(field.Field().String())
}

type Assignment struct {
	RoleName string `yaml:"roleName" validate:"required"`
}

type EligibilitySchedule struct {
	EndDateTime              string `yaml:"endDateTime" validate:"datetime"`
	RoleManagementPolicyName string `yaml:"roleManagementPolicyName"`
	RoleName                 string `yaml:"roleName" validate:"required"`
}

type eligibilitySchedules struct {
	Subscription   []*EligibilitySchedule `yaml:"subscription"`
	ResourceGroups []*EligibilitySchedule `yaml:"resourceGroups" validate:"required"`
	Resources      []*EligibilitySchedule
}

type activeAssignments struct {
	Subscription   []*Assignment `yaml:"subscription"`
	ResourceGroups []*Assignment `yaml:"resourceGroups" validate:"required"`
	Resources      []*Assignment
}

type groupData struct {
	Active   activeAssignments    `yaml:"activeAssignments"`
	Eligible eligibilitySchedules `yaml:"eligibilitySchedules"`
}

func ValidateConfiguration(path string) (valid bool, err error) {
	var validate = validator.New()
	err = validate.RegisterValidation("datetime", validateDateTime)
	if err != nil {
		output.Println("failed to register validation, err: %v", err)
	}

	filename, _ := filepath.Abs(path)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var g groupData

	err = yaml.Unmarshal(yamlFile, &g)
	if err != nil {
		panic(err)
	}

	output.Println(g.Active.Subscription[0].RoleName)
	output.Println(g.Eligible.Subscription[0].RoleManagementPolicyName)

	err = validate.Struct(g)
	if err != nil {
		return false, err
	}

	return true, nil
}
