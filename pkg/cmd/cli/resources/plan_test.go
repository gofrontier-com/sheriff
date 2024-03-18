package resources

import (
	"testing"
)

func TestNewCmdPlanResources(t *testing.T) {
	cmd := NewCmdPlanResources()

	if cmd.Use != "resources" {
		t.Errorf("Use is not correct")
	}
}
