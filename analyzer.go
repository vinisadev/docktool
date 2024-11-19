package main

import (
	"os"
	"path/filepath"
	"strings"
)

func NewProjectAnalyzer(path string) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath:  path,
		files:     make([]string, 0),
		envConfig: NewEnvConfig(),
	}
}

func (pa *ProjectAnalyzer) Analyze() error {
	return filepath.Walk(pa.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(pa.rootPath, path)
			if err != nil {
				return err
			}
			pa.files = append(pa.files, relPath)
		}
		return nil
	})
}

func (pa *ProjectAnalyzer) hasFile(filename string) bool {
	for _, file := range pa.files {
		if strings.EqualFold(filepath.Base(file), filename) {
			return true
		}
	}
	return false
}

func (pa *ProjectAnalyzer) hasAnyFile(extensions []string) bool {
	for _, file := range pa.files {
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(file), ext) {
				return true
			}
		}
	}
	return false
}
