package resources

import (
	"fmt"
	"strings"

	"github.com/gofrontier-com/go-utils/output"
)

var (
	configDir      string
	planOnly       bool
	subscriptionId string
)

var (
	aliases []string = []string{"azurerm", "azure-resources"}
	use     string   = "resources"
)

func printHeader(action string, configDir string, scope *string) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	builder.WriteString(fmt.Sprintf("Action           | %s\n", action))
	builder.WriteString(fmt.Sprintf("Mode             | %s\n", "Azure Resources"))
	builder.WriteString(fmt.Sprintf("Config path      | %s\n", configDir))
	if scope != nil {
		builder.WriteString(fmt.Sprintf("Subscription Id  | %s\n", *scope))
	}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	output.PrintlnInfo(builder.String())
}
