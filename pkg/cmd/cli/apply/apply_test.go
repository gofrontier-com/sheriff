package apply

import (
	"testing"
)

func TestNewCmdApply(t *testing.T) {
	cmd := NewCmdApply()

	if cmd.Use != "apply" {
		t.Errorf("Use is not correct")
	}
}
