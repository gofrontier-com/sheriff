package validate

import (
	"testing"
)

func TestNewCmdValidateAzureRbac(t *testing.T) {
	cmd := NewCmdValidateAzureRbac()

	if cmd.Use != "azurerbac" {
		t.Errorf("Use is not correct")
	}
}
