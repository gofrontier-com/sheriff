package plan

import "testing"

func TestNewCmdPlan(t *testing.T) {
	cmd := NewCmdPlan()

	if cmd.Use != "plan" {
		t.Errorf("Use is not correct")
	}
}
