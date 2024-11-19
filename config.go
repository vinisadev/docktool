package main

import "fmt"

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
