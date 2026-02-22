package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Glazel version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Glazel v0.1.0")
	},
}
