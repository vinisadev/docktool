package main

import "io/ioutil"

func (dc *DockerConfig) SaveToFile() error {
	var content string
	if dc.isCompose {
		content = dc.GenerateComposeContent()
		return ioutil.WriteFile("docker-compose.yml", []byte(content), 0644)
	}

	content = dc.GenerateDockerfileContent()
	return ioutil.WriteFile("Dockerfile", []byte(content), 0644)
}
