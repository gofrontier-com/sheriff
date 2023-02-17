package version

import (
	"testing"
)

func TestNewCmdVersion(t *testing.T) {
	cmd := NewCmdVersion("0.0.0", "commitid", "date")

	if cmd.Use != "version" {
		t.Errorf("Use is not correct")
	}
}
