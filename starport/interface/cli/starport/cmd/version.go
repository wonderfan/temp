package starportcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhigui-projects/zeus-onestop/starport/internal/version"
)

// NewVersion creates a new version command to show starport's version.
func NewVersion() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.Long())
		},
	}
	return c
}
