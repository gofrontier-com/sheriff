package validate

import (
	"testing"
)

func TestNewCmdValidateAzureRm(t *testing.T) {
	cmd := NewCmdValidateAzureRm()

	if cmd.Use != "azurerm" {
		t.Errorf("Use is not correct")
	}
}
