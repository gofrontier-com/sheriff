package plan

import (
	"testing"
)

func TestNewCmdPlanAzureRm(t *testing.T) {
	cmd := NewCmdPlanAzureRm()

	if cmd.Use != "azurerm" {
		t.Errorf("Use is not correct")
	}
}
