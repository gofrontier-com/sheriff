package sheriff

import (
	"testing"
)

func TestNewCmdRoot(t *testing.T) {
	cmd := NewRootCmd("0.0.0", "commitid", "date")

	if cmd.Use != "sheriff" {
		t.Errorf("Use is not correct")
	}
}
