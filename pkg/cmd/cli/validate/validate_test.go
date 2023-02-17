package validate

import (
	"testing"
)

func TestNewCmdValidate(t *testing.T) {
	cmd := NewCmdValidate()

	if cmd.Use != "validate" {
		t.Errorf("Use is not correct")
	}
}
