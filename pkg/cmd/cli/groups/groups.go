package groups

import (
	"fmt"
	"strings"

	"github.com/gofrontier-com/go-utils/output"
)

var (
	configDir string
	planOnly  bool
)

var (
	aliases []string = []string{}
	use     string   = "groups"
)

func printHeader(action string, configDir string) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	builder.WriteString(fmt.Sprintf("Action           | %s\n", action))
	builder.WriteString(fmt.Sprintf("Mode             | %s\n", "Groups"))
	builder.WriteString(fmt.Sprintf("Config path      | %s\n", configDir))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	output.PrintlnInfo(builder.String())
}
