package apply

import (
	"testing"
)

func TestNewCmdApplyAzureRm(t *testing.T) {
	cmd := NewCmdApplyAzureRm()

	if cmd.Use != "azurerm" {
		t.Errorf("Use is not correct")
	}
}
