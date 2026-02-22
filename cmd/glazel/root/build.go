package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build a target using Glazel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		fmt.Printf("Building %s (dispatch coming in Step 3)...\n", target)
	},
}
