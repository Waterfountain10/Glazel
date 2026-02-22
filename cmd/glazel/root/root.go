package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var redisAddr string

var rootCmd = &cobra.Command{
	Use:   "glazel",
	Short: "Glazel distributed build system",
	Long:  "Glazel is a lightweight distributed build cache and orchestrator.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&redisAddr,
		"redis",
		"localhost:6379",
		"Redis address",
	)

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(versionCmd)
}
