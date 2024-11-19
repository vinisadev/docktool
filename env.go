package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func NewEnvConfig() *EnvConfig {
	return &EnvConfig{
		variables: make(map[string]string),
		secrets:   make([]string, 0),
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
