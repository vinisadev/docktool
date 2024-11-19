package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ProjectAnalyzer struct {
	rootPath string
	files    []string
}

type DockerConfig struct {
	baseImage   string
	ports       []string
	commands    []string
	environment map[string]string
	isCompose   bool
	services    map[string]ServiceConfig
}

type ServiceConfig struct {
	name         string
	baseImage    string
	ports        []string
	commands     []string
	environment  map[string]string
	dependencies []string
}

func main() {
	composeFlag := flag.Bool("compose", false, "Generate docker-compose.yml instead of Dockerfile")
	pathFlag := flag.String("path", ".", "Path to the project directory")
	flag.Parse()

	analyzer := NewProjectAnalyzer(*pathFlag)
	if err := analyzer.Analyze(); err != nil {
		fmt.Printf("Error analyzing project: %v\n", err)
		os.Exit(1)
	}

	config := analyzer.GenerateDockerConfig(*composeFlag)
	if err := config.SaveToFile(); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}
}

func NewProjectAnalyzer(path string) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath: path,
		files:    make([]string, 0),
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

func (pa *ProjectAnalyzer) GenerateDockerConfig(isCompose bool) *DockerConfig {
	config := &DockerConfig{
		isCompose:   isCompose,
		environment: make(map[string]string),
		services:    make(map[string]ServiceConfig),
	}

	// Detect project type and set appropriate configurations
	if pa.hasFile("package.json") {
		config.configureNodeJS()
	} else if pa.hasFile("requirements.txt") || pa.hasFile("Pipfile") {
		config.configurePython()
	} else if pa.hasFile("go.mod") {
		config.configureGo()
	} else {
		config.configureGeneric()
	}

	return config
}

func (pa *ProjectAnalyzer) hasFile(filename string) bool {
	for _, file := range pa.files {
		if strings.EqualFold(filepath.Base(file), filename) {
			return true
		}
	}
	return false
}

func (dc *DockerConfig) configureNodeJS() {
	dc.baseImage = "node:18-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY package*.json ./",
		"RUN npm install",
		"COPY . .",
		"EXPOSE 3000",
		"CMD [\"npm\", \"start\"]",
	}
	dc.ports = []string{"3000:3000"}
}

func (dc *DockerConfig) configurePython() {
	dc.baseImage = "python:3.9-slim"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY requirements.txt .",
		"RUN pip install --no-cache-dir -r requirements.txt",
		"COPY . .",
		"EXPOSE 8000",
		"CMD [\"python\", \"app.py\"]",
	}
	dc.ports = []string{"8000:8000"}
}

func (dc *DockerConfig) configureGo() {
	dc.baseImage = "golang:1.20-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY go.* .",
		"RUN go mod download",
		"COPY . .",
		"RUN go build -o main .",
		"EXPOSE 8080",
		"CMD [\"./main\"]",
	}
	dc.ports = []string{"8080:8080"}
}

func (dc *DockerConfig) configureGeneric() {
	dc.baseImage = "ubuntu:latest"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY . .",
		"CMD [\"/bin/bash\"]",
	}
}

func (dc *DockerConfig) SaveToFile() error {
	var content string
	if dc.isCompose {
		content = dc.generateComposeContent()
		return ioutil.WriteFile("docker-compose.yml", []byte(content), 0644)
	}

	content = dc.generateDockerfileContent()
	return ioutil.WriteFile("Dockerfile", []byte(content), 0644)
}

func (dc *DockerConfig) generateDockerfileContent() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("FROM %s\n\n", dc.baseImage))

	for _, cmd := range dc.commands {
		sb.WriteString(cmd + "\n")
	}

	return sb.String()
}

func (dc *DockerConfig) generateComposeContent() string {
	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\nservices:\n")
	sb.WriteString("  app:\n")
	sb.WriteString(fmt.Sprintf("    image: %s\n", dc.baseImage))
	sb.WriteString("    build:\n")
	sb.WriteString("      context: .\n")

	if len(dc.ports) > 0 {
		sb.WriteString("    ports:\n")
		for _, port := range dc.ports {
			sb.WriteString(fmt.Sprintf("      - %s\n", port))
		}
	}

	return sb.String()
}
