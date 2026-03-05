//go:build !cli

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type FileInfo struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	SizeStr string `json:"sizeStr"`
}

func (a *App) SelectFile() (*FileInfo, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select mbox file",
		Filters: []runtime.FileFilter{
			{DisplayName: "Mbox Files (*.mbox)", Pattern: "*.mbox"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	return a.getFileInfo(path)
}

func (a *App) GetFileInfo(path string) (*FileInfo, error) {
	return a.getFileInfo(path)
}

func (a *App) getFileInfo(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Path:    path,
		Name:    info.Name(),
		Size:    info.Size(),
		SizeStr: formatSize(info.Size()),
	}, nil
}

func (a *App) SelectOutputDir() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select output directory",
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) Split(filePath string, maxSizeStr string, outputDir string) (*SplitResult, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil, fmt.Errorf("file path is required")
	}

	maxSize, err := parseMaxSize(maxSizeStr)
	if err != nil || maxSize <= 0 {
		return nil, fmt.Errorf("invalid max size")
	}

	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		base := strings.TrimSuffix(filePath, filepath.Ext(filePath))
		outputDir = base + "_split"
	}

	sp := &splitter{
		inputPath: filePath,
		outputDir: outputDir,
		maxSize:   maxSize,
	}

	return sp.run()
}
