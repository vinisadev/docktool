# DockTool

DockTool is a command-line tool that automatically analyzes your project directory and generates appropriate Docker configurations. It supports multiple project types and can create both Dockerfiles and docker-compose.yml files.

## Features

- 🔍 Automatic project type detection
- 📦 Supports multiple project types (Node.js, Python, Go)
- 🔄 Generates both Dockerfile and docker-compose.yml
- 🛠️ Smart defaults based on project type
- 💪 Easily extensible for additional project types

## Supported Project Types

| Project Type | Detection Method | Base Image | Default Port |
|-------------|------------------|------------|--------------|
| Node.js | package.json | node:18-alpine | 3000 |
| Python | requirements.txt/Pipfile | python:3.9-slim | 8000 |
| Go | go.mod | golang:1.20-alpine | 8080 |
| Generic | fallback | ubuntu:latest | none |

## Installation

### Prerequisites

- Go 1.18 or higher
- Git (for cloning the repository)

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/vinisadev/docktool.git
cd docktool
```

2. Build the binary:
```bash
go build -o docktool
```

3. (Optional) Install globally:
```bash
go install
```

## Usage

### Basic Usage

Generate a Dockerfile for the current directory:
```bash
./docktool
```

Generate a docker-compose.yml file:
```bash
./docktool -compose
```

Analyze a specific project directory:
```bash
./docktool -path /path/to/project
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-compose` | Generate docker-compose.yml instead of Dockerfile | false |
| `-path` | Specify the project directory to analyze | "." (current directory) |

## Generated Configurations

### Node.js Project Example
```dockerfile
FROM node:18-alpine

WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

### Python Project Example
```dockerfile
FROM python:3.9-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "app.py"]
```

### Go Project Example
```dockerfile
FROM golang:1.20-alpine

WORKDIR /app
COPY go.* .
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]
```

## Contributing

Contributions are welcome! Here are some ways you can contribute:

1. Add support for additional project types
2. Improve detection logic
3. Add more configuration options
4. Report bugs and suggest features
5. Improve documentation

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Troubleshooting

### Common Issues

1. **Permission Denied**: If you get a permission denied error when generating files, ensure you have write permissions in the target directory.
   ```bash
   sudo chmod +w /path/to/directory
   ```

2. **Project Type Not Detected**: If your project type isn't detected, check if the required files (package.json, requirements.txt, etc.) are in the root directory.

3. **Build Errors**: Ensure you have Go 1.18 or higher installed:
   ```bash
   go version
   ```

### Getting Help

If you encounter any issues:

1. Check the existing GitHub issues
2. Create a new issue with:
   - Your operating system
   - Go version
   - Project type
   - Error message
   - Steps to reproduce

## Roadmap

Future improvements planned:

- [ ] Support for more project types (Java, Ruby, PHP)
- [ ] Custom template support
- [ ] Multi-service docker-compose configurations
- [ ] Configuration file support
- [ ] Interactive mode
- [ ] Environment variable detection
- [ ] Database service auto-configuration

## Acknowledgments

- Docker documentation
- Go community
- All contributors