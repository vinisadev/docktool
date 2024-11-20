package main

import (
	"fmt"
	"os"

	"github.com/vinisadev/docktool/pkg/detect"
	"github.com/vinisadev/docktool/pkg/generate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: docktool <project-directory>")
		os.Exit(1)
	}

	projectDir := os.Args[1]
	projectType, err := detect.DetectProjectType(projectDir)
	if err != nil {
		fmt.Printf("Error detecting project type: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Detect project type: %s\n", projectType)

	err = generate.GenerateDockerFiles(projectDir, projectType)
	if err != nil {
		fmt.Printf("Error generating Docker files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Docker files generated successfully!")
}
