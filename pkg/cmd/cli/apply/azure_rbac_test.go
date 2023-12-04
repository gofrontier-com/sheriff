package apply

import (
	"testing"
)

func TestNewCmdApplyAzureRbac(t *testing.T) {
	cmd := NewCmdApplyAzureRbac()

	if cmd.Use != "azurerbac" {
		t.Errorf("Use is not correct")
	}
}
