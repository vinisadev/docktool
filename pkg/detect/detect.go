package detect

import (
	"errors"
	"os"
)

func DetectProjectType(projectDir string) (string, error) {
	files, err := os.ReadDir(projectDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		switch file.Name() {
		case "package.json":
			return "nodejs", nil
		case "pom.xml":
			return "java", nil
		case "requirements.txt":
			return "python", nil
		case "Gemfile":
			return "ruby", nil
		case "go.mod":
			return "go", nil
		case "composer.json":
			return "php", nil
		}
	}

	return "", errors.New("unable to detect project type")
}
