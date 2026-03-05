//go:build cli

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	maxSizeStr := flag.String("max-size", "100MB", "Maximum size per output file (e.g. 50MB, 1GB, 500KB)")
	outputDir := flag.String("output", "", "Output directory (default: <input>_split)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mbox-splitter [options] <input.mbox>\n\nSplits a large mbox file into smaller mbox files.\n\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  mbox-splitter -max-size 50MB mail.mbox\n")
		fmt.Fprintf(os.Stderr, "  mbox-splitter -max-size 1GB -output /tmp/split mail.mbox\n")
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("mbox-splitter %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)

	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access input file: %v\n", err)
		os.Exit(1)
	}

	maxSize, err := parseMaxSize(*maxSizeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid max-size: %v\n", err)
		os.Exit(1)
	}
	if maxSize <= 0 {
		fmt.Fprintf(os.Stderr, "Error: max-size must be positive\n")
		os.Exit(1)
	}

	outDir := *outputDir
	if outDir == "" {
		base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
		outDir = base + "_split"
	}

	sp := &splitter{
		inputPath: inputPath,
		outputDir: outDir,
		maxSize:   maxSize,
	}

	fmt.Printf("Input: %s\n", inputPath)
	fmt.Printf("Max output size: %s\n", formatSize(maxSize))
	fmt.Printf("Output directory: %s\n\n", outDir)

	result, err := sp.run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, f := range result.Files {
		fmt.Printf("  Wrote %s (%s, %d messages)\n", f.Name, f.SizeStr, f.Messages)
	}
	fmt.Printf("\nDone. Split %d messages into %d files.\n", result.TotalMsgs, len(result.Files))
}
