package generate

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateDockerFiles(projectDir, projectType string) error {
	envVars, err := readEnvFile(filepath.Join(projectDir, ".env"))
	if err != nil {
		return err
	}

	dockerfileContent, err := getDockerfileContent(projectType, envVars)
	if err != nil {
		return err
	}

	dockerComposeContent, err := getDockerComposeContent(projectType, envVars)
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(projectDir, "Dockerfile"), dockerfileContent)
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(projectDir, "docker-compose.yml"), dockerComposeContent)
	if err != nil {
		return err
	}

	return nil
}

func readEnvFile(filePath string) (map[string]string, error) {
	envVars := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return envVars, nil // Return an empty map if the file does not exist
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}

func getDockerfileContent(projectType string, envVars map[string]string) (string, error) {
	envSection := ""
	for key, value := range envVars {
		envSection += fmt.Sprintf("ENV %s=%s\n", key, value)
	}

	switch projectType {
	case "nodejs":
		return fmt.Sprintf(`FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
%s
EXPOSE 3000
CMD ["node", "index.js"]
    `, envSection), nil
	case "java":
		return fmt.Sprintf(`FROM openjdk:11-jre-slim
WORKDIR /app
COPY . .
%s
RUN ./mvnw clean package
CMD ["java", "-jar", "target/myapp.jar"]
    `, envSection), nil
	case "python":
		return fmt.Sprintf(`FROM python:3.8-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
%s
CMD ["python", "app.py"]
    `, envSection), nil
	case "ruby":
		return fmt.Sprintf(`FROM ruby:2.7
WORKDIR /app
COPY Gemfile* .
RUN bundle install
COPY . .
%s
CMD ["ruby", "app.rb"]
    `, envSection), nil
	case "go":
		return fmt.Sprintf(`FROM golang:1.18-alpine
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
%s
RUN go build -o /dockerapp
EXPOSE 8080
CMD ["/dockerapp"]
    `, envSection), nil
	case "php":
		return fmt.Sprintf(`FROM php:7.4-cli
WORKDIR /app
COPY composer.json composer.lock ./
RUN curl -sS https://getcomposer.org/install | php -- --install-dir=/usr/local/bin --filename=composer
RUN comoser install
COPY . .
%s
CMD ["php", "./your-script.php"]
`, envSection), nil
	default:
		return "", fmt.Errorf("unsupported project type: %s", projectType)
	}
}

func getDockerComposeContent(projectType string, envVars map[string]string) (string, error) {
	envSection := ""
	if len(envVars) > 0 {
		envSection = "    environment:\n"
		for key, value := range envVars {
			envSection += fmt.Sprintf("       - %s=%s\n", key, value)
		}
	}

	switch projectType {
	case "nodejs":
		return fmt.Sprintf(`version: '3'
services:
  app:
    build: .
    ports:
      - "3000:3000"
%s
    `, envSection), nil
	case "java":
		return fmt.Sprintf(`version: '3'
services:
  app:
    build: .
    ports:
      - "8080:8080"
%s
    `, envSection), nil
	case "python":
		return fmt.Sprintf(`version: '3'
services:
  app:
    build: .
    ports:
      - "5000:5000"
%s
    `, envSection), nil
	case "ruby":
		return fmt.Sprintf(`version: '3'
services:
  app:
    build: .
    ports:
      - "3000:3000"
%s
    `, envSection), nil
	case "go":
		return fmt.Sprintf(`version: '3'
services:
  app:
    build: .
    ports:
      - "8080:8080"
%s
    `, envSection), nil
	case "php":
		return fmt.Sprintf(`version: '3'
services:
  app:
    build: .
    ports:
      - "8000:8000
%s
    `, envSection), nil
	default:
		return "", fmt.Errorf("unsupported project type: %s", projectType)
	}
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
