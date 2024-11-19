package main

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
