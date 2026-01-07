//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s VERSION\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\nUpdate the version string in cmd/droopy/main.go\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	version := flag.Arg(0)
	filePath := "cmd/droopy/main.go"

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`version\s+string\s*=\s*"[^"]*"`)
	updated := re.ReplaceAll(data, []byte(fmt.Sprintf(`version     string = "%s"`, version)))

	err = os.WriteFile(filePath, updated, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated version to %s\n", version)
}
