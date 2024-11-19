package main

import (
	"flag"
	"fmt"
	"os"
)

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
