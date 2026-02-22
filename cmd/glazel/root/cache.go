package root

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache inspection commands",
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Run: func(cmd *cobra.Command, args []string) {

		root := ".glazel/cas/obj"

		totalSize := int64(0)
		count := 0

		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				count++
				totalSize += info.Size()
			}
			return nil
		})

		fmt.Printf("\n%s%s>>> GLAZEL CACHE%s\n\n", bold, blue, reset)
		fmt.Printf("%sObjects:%s %d\n", dim, reset, count)
		fmt.Printf("%sTotal Size:%s %.2f KB\n\n", dim, reset,
			float64(totalSize)/1024.0)
	},
}

func init() {
	cacheCmd.AddCommand(cacheStatsCmd)
	rootCmd.AddCommand(cacheCmd)
}
