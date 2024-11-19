package main

import (
	"fmt"
	"strings"
)

func (dc *DockerConfig) GenerateDockerfileContent() string {
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

func (dc *DockerConfig) GenerateComposeContent() string {
	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\n")

	// Add top-level secrets if any are defined
	if len(dc.services) > 0 && dc.hasSecrets() {
		sb.WriteString("secrets:\n")
		for key := range dc.environment {
			if dc.isSecretVariable(key) {
				sb.WriteString(fmt.Sprintf("	%s:\n", strings.ToLower(key)))
				sb.WriteString("		file: .env\n")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("services:\n")
	sb.WriteString("	app:\n")
	sb.WriteString(fmt.Sprintf("		image: %s\n", dc.baseImage))
	sb.WriteString("		build:\n")
	sb.WriteString("			context: .\n")

	if len(dc.ports) > 0 {
		sb.WriteString("		ports:\n")
		for _, port := range dc.ports {
			sb.WriteString(fmt.Sprintf("			- %s\n", port))
		}
	}

	// Environment variable handling
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
			sb.WriteString("		environment:\n")
			for _, env := range envVars {
				sb.WriteString(fmt.Sprintf("			- %s\n", env))
			}
		}

		// Secrets handling
		if len(secrets) > 0 {
			sb.WriteString("		secrets:\n")
			for _, secret := range secrets {
				sb.WriteString(fmt.Sprintf("			- %s\n", strings.ToLower(secret)))
			}
		}
	}

	return sb.String()
}
