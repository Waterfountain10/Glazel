package root

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Waterfountain10/glazel/internal/api"
	"github.com/spf13/cobra"
)

func collectCppFiles(path string) ([]string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	out := []string{}
	if !st.IsDir() {
		if strings.HasSuffix(path, ".cpp") {
			return []string{path}, nil
		}
		return nil, fmt.Errorf("not a .cpp file: %s", path)
	}

	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(p, ".cpp") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

var (
	compiler string
	outName  string
	cxxflags string
)

var buildCmd = &cobra.Command{
	Use:   "build [path-or-file]",
	Short: "Distributed compile with cache (linking next milestone)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		files, err := collectCppFiles(args[0])
		if err != nil {
			fmt.Println("collect files error:", err)
			return
		}
		if len(files) == 0 {
			fmt.Println("no .cpp files found")
			return
		}

		flags := []string{}
		if strings.TrimSpace(cxxflags) != "" {
			flags = strings.Fields(cxxflags)
		}

		req := api.BuildRequest{
			Files:    files,
			Out:      outName,
			Compiler: compiler,
			CxxFlags: flags,
		}

		b, _ := json.Marshal(req)
		t0 := time.Now()
		httpResp, err := http.Post(serverAddr+"/v1/build", "application/json", bytes.NewReader(b))
		if err != nil {
			fmt.Println("build request failed:", err)
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != 200 {
			buf, _ := io.ReadAll(httpResp.Body)
			fmt.Println(string(buf))
			return
		}

		var resp api.BuildResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
			fmt.Println("bad response:", err)
			return
		}

		fmt.Printf("\n%s%s>>> GLAZEL BUILD%s\n\n", bold, blue, reset)
		fmt.Printf("%sFiles:%s %d   %sServer:%s %s\n\n", dim, reset, len(files), dim, reset, serverAddr)

		fmt.Printf("%-28s %-10s %-8s %-6s\n", "FILE", "WORKER", "STATUS", "HASH")
		fmt.Println("--------------------------------------------------------------")

		for _, row := range resp.Rows {
			statusColor := yell
			if row.Status == "HIT" {
				statusColor = green
			}
			worker := row.WorkerID
			if worker == "" {
				worker = "-"
			}
			fmt.Printf("%-28s %-10s %s%-8s%s %-6s\n",
				shorten(row.File, 28),
				worker,
				statusColor, row.Status, reset,
				row.Hash4,
			)
		}

		dt := time.Since(t0)
		total := resp.CacheHits + resp.CacheMisses
		hitRate := 0
		if total > 0 {
			hitRate = int((100 * resp.CacheHits) / total)
		}

		fmt.Printf("\n%sCache:%s %d hit / %d miss  (%d%% hit rate)\n", dim, reset, resp.CacheHits, resp.CacheMisses, hitRate)
		fmt.Printf("%sTime:%s  %s\n\n", dim, reset, dt.Round(time.Millisecond))
		fmt.Printf("%sOutput:%s %s\n", dim, reset, resp.OutPath)
	},
}

func init() {
	buildCmd.Flags().StringVar(&compiler, "compiler", "g++", "C++ compiler (e.g. g++, clang++)")
	buildCmd.Flags().StringVar(&outName, "out", "a.out", "Output name (linking next milestone)")
	buildCmd.Flags().StringVar(&cxxflags, "cxxflags", "-O2 -std=c++20", "C++ flags (quoted string)")
}

func shorten(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return "..." + s[len(s)-(max-3):]
}
