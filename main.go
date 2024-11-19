package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ProjectAnalyzer struct {
	rootPath  string
	files     []string
	envConfig *EnvConfig
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

type EnvConfig struct {
	variables map[string]string
	secrets   []string
}

func NewEnvConfig() *EnvConfig {
	return &EnvConfig{
		variables: make(map[string]string),
		secrets:   make([]string, 0),
	}
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
		rootPath:  path,
		files:     make([]string, 0),
		envConfig: NewEnvConfig(),
	}
}

func (pa *ProjectAnalyzer) detectEnvironmentVariables() error {
	envFiles := []string{".env", ".env.example", ".env.template", ".env.default"}

	// Check for environment files
	for _, envFile := range envFiles {
		envPath := filepath.Join(pa.rootPath, envFile)
		if _, err := os.Stat(envPath); err == nil {
			if err := pa.parseEnvFile(envPath); err != nil {
				return fmt.Errorf("error parsing %s: %v", envFile, err)
			}
			break
		}
	}

	// Look for environment variables in other common config files
	pa.detectFrameworkEnvironmentVars()

	return nil
}

func (pa *ProjectAnalyzer) parseEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	commentPattern := regexp.MustCompile(`^\s*#`)
	variablePattern := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || commentPattern.MatchString(line) {
			continue
		}

		matches := variablePattern.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := strings.Trim(matches[2], `"'`)

			pa.envConfig.variables[key] = value

			// Detect if this might be a secret
			if pa.isLikelySecret(key, value) {
				pa.envConfig.secrets = append(pa.envConfig.secrets, key)
			}
		}
	}

	return scanner.Err()
}

func (pa *ProjectAnalyzer) detectFrameworkEnvironmentVars() {
	switch {
	case pa.hasFile("package.json"):
		pa.detectNodeEnvironmentVars()
	case pa.hasFile("requirements.txt"):
		pa.detectPythonEnvironmentVars()
	case pa.hasFile("composer.json"):
		pa.detectPHPEnvironmentVars()
	case pa.hasFile("Gemfile"):
		pa.detectRubyEnvironmentVars()
	}
}

func (pa *ProjectAnalyzer) isLikelySecret(key, value string) bool {
	secretPatterns := []string{
		"(?i)password",
		"(?i)secret",
		"(?i)token",
		"(?i)key",
		"(?i)auth",
		"(?i)credential",
		"(?i)cert",
	}

	for _, pattern := range secretPatterns {
		if matched, _ := regexp.MatchString(pattern, key); matched {
			return true
		}
	}

	return false
}

func (pa *ProjectAnalyzer) detectNodeEnvironmentVars() {
	commonVars := map[string]string{
		"NODE_ENV": "production",
		"PORT":     "3000",
	}

	for k, v := range commonVars {
		if _, exists := pa.envConfig.variables[k]; !exists {
			pa.envConfig.variables[k] = v
		}
	}
}

func (pa *ProjectAnalyzer) detectPythonEnvironmentVars() {
	commonVars := map[string]string{
		"PYTHONPATH":             "/app",
		"FLASK_ENV":              "production",
		"DJANGO_SETTINGS_MODULE": "project.settings.production",
	}

	for k, v := range commonVars {
		if _, exists := pa.envConfig.variables[k]; !exists {
			pa.envConfig.variables[k] = v
		}
	}
}

func (pa *ProjectAnalyzer) detectRubyEnvironmentVars() {
	commonVars := map[string]string{
		"RAILS_ENV": "production",
		"RACK_ENV":  "production",
	}

	for k, v := range commonVars {
		if _, exists := pa.envConfig.variables[k]; !exists {
			pa.envConfig.variables[k] = v
		}
	}
}

func (pa *ProjectAnalyzer) detectPHPEnvironmentVars() {
	commonVars := map[string]string{
		"APP_ENV":   "production",
		"APP_DEBUG": "false",
	}

	for k, v := range commonVars {
		if _, exists := pa.envConfig.variables[k]; !exists {
			pa.envConfig.variables[k] = v
		}
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

	if err := pa.detectEnvironmentVariables(); err != nil {
		fmt.Printf("Warning: Error detecting environment variables: %v\n", err)
	}

	for k, v := range pa.envConfig.variables {
		config.environment[k] = v
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

	if len(dc.environment) > 0 {
		sb.WriteString("# Build arguments\n")
		for key := range dc.environment {
			sb.WriteString(fmt.Sprintf("ARG %s\n", key))
		}
		sb.WriteString("\n")
	}

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
	sb.WriteString("version: '3.8'\n\n")

	// Add top-level secrets if any are defined
	if len(dc.services) > 0 && dc.hasSecrets() {
		sb.WriteString("secrets:\n")
		for key := range dc.environment {
			if dc.isSecretVariable(key) {
				sb.WriteString(fmt.Sprintf("  %s:\n", strings.ToLower(key)))
				sb.WriteString("    file: .env\n")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("services:\n")
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

	// Environment variables handling
	if len(dc.environment) > 0 {
		// Regular environment variables
		var envVars, secrets []string
		for key, value := range dc.environment {
			if dc.isSecretVariable(key) {
				secrets = append(secrets, key)
			} else {
				envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
			}
		}

		if len(envVars) > 0 {
			sb.WriteString("    environment:\n")
			for _, env := range envVars {
				sb.WriteString(fmt.Sprintf("      - %s\n", env))
			}
		}

		// Secrets handling
		if len(secrets) > 0 {
			sb.WriteString("    secrets:\n")
			for _, secret := range secrets {
				sb.WriteString(fmt.Sprintf("      - %s\n", strings.ToLower(secret)))
			}
		}
	}

	return sb.String()
}

func (dc *DockerConfig) isSecretVariable(key string) bool {
	secretPatterns := []string{
		"(?i)password",
		"(?i)secret",
		"(?i)token",
		"(?i)key",
		"(?i)auth",
		"(?i)credential",
		"(?i)cert",
	}

	for _, pattern := range secretPatterns {
		if matched, _ := regexp.MatchString(pattern, key); matched {
			return true
		}
	}
	return false
}

func isSecret(key string, dc *DockerConfig) bool {
	secretPatterns := []string{
		"(?i)password",
		"(?i)secret",
		"(?i)token",
		"(?i)key",
		"(?i)auth",
		"(?i)credential",
		"(?i)cert",
	}

	for _, pattern := range secretPatterns {
		if matched, _ := regexp.MatchString(pattern, key); matched {
			return true
		}
	}
	return false
}

func hasSecret(dc *DockerConfig) bool {
	for key := range dc.environment {
		if isSecret(key, dc) {
			return true
		}
	}
	return false
}

func (dc *DockerConfig) hasSecrets() bool {
	for key := range dc.environment {
		if dc.isSecretVariable(key) {
			return true
		}
	}
	return false
}
