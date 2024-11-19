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

	fmt.Println("Docker configuration generated successfully!")
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

	switch {
	case pa.hasFile("package.json"):
		pa.configureNodeJS(config)
	case pa.hasFile("requirements.txt") || pa.hasFile("Pipfile"):
		pa.configurePython(config)
	case pa.hasFile("go.mod"):
		pa.configureGo(config)
	case pa.hasFile("pom.xml") || pa.hasFile("build.gradle"):
		pa.configureJava(config)
	case pa.hasFile("Gemfile"):
		pa.configureRuby(config)
	case pa.hasFile("composer.json") || pa.hasAnyFile([]string{".php"}):
		pa.configurePHP(config)
	default:
		pa.configureGeneric(config)
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

func (pa *ProjectAnalyzer) configureNodeJS(dc *DockerConfig) {
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

func (pa *ProjectAnalyzer) configurePython(dc *DockerConfig) {
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

func (pa *ProjectAnalyzer) configureGo(dc *DockerConfig) {
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

func (pa *ProjectAnalyzer) configureJava(dc *DockerConfig) {
	if pa.hasFile("pom.xml") {
		pa.configureMavenJava(dc)
	} else {
		pa.configureGradleJava(dc)
	}
}

func (pa *ProjectAnalyzer) configureMavenJava(dc *DockerConfig) {
	dc.baseImage = "eclipse-temurin:17-jdk-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY pom.xml .",
		"COPY .mvn .mvn",
		"COPY mvnw .",
		"RUN chmod +x mvnw",
		"RUN ./mvnw dependency:go-offline",
		"COPY src src",
		"RUN ./mvnw package -DskipTests",
		"EXPOSE 8080",
		"CMD [\"java\", \"-jar\", \"target/*.jar\"]",
	}
	dc.ports = []string{"8080:8080"}
}

func (pa *ProjectAnalyzer) configureGradleJava(dc *DockerConfig) {
	dc.baseImage = "eclipse-temurin:17-jdk-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY build.gradle settings.gradle ./",
		"COPY gradle gradle",
		"COPY gradlew .",
		"RUN chmod +x gradlew",
		"RUN ./gradlew dependencies",
		"COPY src src",
		"RUN ./gradlew build -x test",
		"EXPOSE 8080",
		"CMD [\"java\", \"-jar\", \"build/libs/*.jar\"]",
	}
	dc.ports = []string{"8080:8080"}
}

func (pa *ProjectAnalyzer) configureRuby(dc *DockerConfig) {
	dc.baseImage = "ruby:3.2-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY Gemfile Gemfile.lock ./",
		"RUN apk add --no-cache build-base postgresql-dev",
		"RUN bundle install",
		"COPY . .",
		"EXPOSE 3000",
		"CMD [\"bundle\", \"exec\", \"rails\", \"server\", \"-b\", \"0.0.0.0\"]",
	}
	dc.ports = []string{"3000:3000"}
	dc.environment = map[string]string{
		"RAILS_ENV": "production",
	}
}

func (pa *ProjectAnalyzer) configurePHP(dc *DockerConfig) {
	dc.baseImage = "php:8.2-apache"
	dc.commands = []string{
		"WORKDIR /var/www/html",
		"RUN apt-get update && apt-get install -y \\\n" +
			"    libzip-dev \\\n" +
			"    zip \\\n" +
			"    && docker-php-ext-install zip pdo pdo_mysql",
		"COPY --from=composer:latest /usr/bin/composer /usr/bin/composer",
		"COPY composer.* ./",
		"RUN composer install --no-dev --no-scripts --no-autoloader",
		"COPY . .",
		"RUN composer dump-autoload --optimize",
		"RUN chown -R www-data:www-data /var/www/html",
		"EXPOSE 80",
	}
	dc.ports = []string{"80:80"}
}

func (pa *ProjectAnalyzer) configureGeneric(dc *DockerConfig) {
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

	if len(dc.environment) > 0 {
		sb.WriteString("\n# Environment variables\n")
		for key, value := range dc.environment {
			sb.WriteString(fmt.Sprintf("ENV %s=%s\n", key, value))
		}
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

	if len(dc.environment) > 0 {
		sb.WriteString("    environment:\n")
		for key, value := range dc.environment {
			sb.WriteString(fmt.Sprintf("      - %s=%s\n", key, value))
		}
	}

	return sb.String()
}
