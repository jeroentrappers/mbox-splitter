package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var fromLineRe = regexp.MustCompile(`^From \S+`)

var dateLayouts = []string{
	time.ANSIC,
	"Mon Jan _2 15:04:05 MST 2006",
	"Mon Jan _2 15:04:05 -0700 2006",
	"Mon Jan  2 15:04:05 2006",
	"Mon Jan  2 15:04:05 MST 2006",
	"Mon Jan  2 15:04:05 -0700 2006",
}

func parseFromDate(line string) (time.Time, bool) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return time.Time{}, false
	}
	dateStr := parts[2]
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, true
		}
	}
	words := strings.Fields(line)
	for i := 2; i < len(words)-3; i++ {
		candidate := strings.Join(words[i:], " ")
		for _, layout := range dateLayouts {
			if t, err := time.Parse(layout, candidate); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func isFromLine(line string) bool {
	return fromLineRe.MatchString(line)
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func parseMaxSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, fmt.Errorf("empty size string")
	}
	s = strings.ToUpper(s)
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(s, "GB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "G"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "G")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "M"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "K"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	s = strings.TrimSpace(s)
	var value float64
	if _, err := fmt.Sscanf(s, "%f", &value); err != nil {
		return 0, fmt.Errorf("invalid size: %s", s)
	}
	return int64(value * float64(multiplier)), nil
}

type FileResult struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	SizeStr  string `json:"sizeStr"`
	Messages int    `json:"messages"`
}

type SplitResult struct {
	Files     []FileResult `json:"files"`
	TotalMsgs int          `json:"totalMessages"`
	OutputDir string       `json:"outputDir"`
}

type splitter struct {
	inputPath string
	outputDir string
	maxSize   int64

	fileIndex     int
	currentFile   *os.File
	currentWriter *bufio.Writer
	currentSize   int64
	firstDate     string
	msgCount      int
	totalMsgs     int
	results       []FileResult
}

func (sp *splitter) closeCurrentFile() error {
	if sp.currentFile == nil {
		return nil
	}
	if err := sp.currentWriter.Flush(); err != nil {
		return err
	}
	name := sp.currentFile.Name()
	size := sp.currentSize
	msgs := sp.msgCount
	if err := sp.currentFile.Close(); err != nil {
		return err
	}
	sp.results = append(sp.results, FileResult{
		Name:     filepath.Base(name),
		Size:     size,
		SizeStr:  formatSize(size),
		Messages: msgs,
	})
	sp.currentFile = nil
	sp.currentWriter = nil
	sp.currentSize = 0
	sp.msgCount = 0
	return nil
}

func (sp *splitter) openNewFile() error {
	sp.fileIndex++
	datePart := sp.firstDate
	if datePart == "" {
		datePart = "unknown"
	}
	baseName := strings.TrimSuffix(filepath.Base(sp.inputPath), filepath.Ext(sp.inputPath))
	fileName := fmt.Sprintf("%s_%s_%03d.mbox", baseName, datePart, sp.fileIndex)
	path := filepath.Join(sp.outputDir, fileName)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	sp.currentFile = f
	sp.currentWriter = bufio.NewWriterSize(f, 256*1024)
	sp.currentSize = 0
	sp.msgCount = 0
	return nil
}

func (sp *splitter) writeLine(line string) error {
	n, err := sp.currentWriter.WriteString(line)
	if err != nil {
		return err
	}
	sp.currentSize += int64(n)
	return nil
}

func (sp *splitter) run() (*SplitResult, error) {
	if err := os.MkdirAll(sp.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.Open(sp.inputPath)
	if err != nil {
		return nil, fmt.Errorf("opening input file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var msgBuf strings.Builder
	var msgFromLine string
	firstMessage := true

	flushMessage := func() error {
		if msgBuf.Len() == 0 {
			return nil
		}
		msgBytes := msgBuf.String()
		msgSize := int64(len(msgBytes))

		if sp.currentFile != nil && sp.currentSize > 0 && sp.currentSize+msgSize > sp.maxSize {
			if err := sp.closeCurrentFile(); err != nil {
				return err
			}
		}

		if sp.currentFile == nil {
			if t, ok := parseFromDate(msgFromLine); ok {
				sp.firstDate = t.Format("2006-01-02")
			}
			if err := sp.openNewFile(); err != nil {
				return err
			}
		}

		if err := sp.writeLine(msgBytes); err != nil {
			return err
		}
		sp.msgCount++
		sp.totalMsgs++
		msgBuf.Reset()
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text() + "\n"
		if isFromLine(line) {
			if !firstMessage {
				if err := flushMessage(); err != nil {
					return nil, err
				}
			}
			firstMessage = false
			msgFromLine = line
			msgBuf.Reset()
		}
		msgBuf.WriteString(line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	if err := flushMessage(); err != nil {
		return nil, err
	}

	if err := sp.closeCurrentFile(); err != nil {
		return nil, err
	}

	return &SplitResult{
		Files:     sp.results,
		TotalMsgs: sp.totalMsgs,
		OutputDir: sp.outputDir,
	}, nil
}
