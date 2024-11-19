package main

func (pa *ProjectAnalyzer) configureNodeJS(dc *DockerConfig) {
	dc.baseImage = "node:18-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY package*.json ./",
		"RUN npm install",
		"COPY . .",
		"EXPOSE 3000",
		"CMD [\"npm\", \"start\"]",
	}
	dc.ports = []string{"3000:3000"}
}

func (pa *ProjectAnalyzer) configurePython(dc *DockerConfig) {
	dc.baseImage = "python:3.9-slim"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY requirements.txt .",
		"RUN pip install --no-cache-dir -r requirements.txt",
		"COPY . .",
		"EXPOSE 8000",
		"CMD [\"python\", \"app.py\"]",
	}
	dc.ports = []string{"8000:8000"}
}

func (pa *ProjectAnalyzer) configureGo(dc *DockerConfig) {
	dc.baseImage = "golang:1.20-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY go.* .",
		"RUN go mod download",
		"COPY . .",
		"RUN go build -o main .",
		"EXPOSE 8080",
		"CMD [\"./main\"]",
	}
	dc.ports = []string{"8080:8080"}
}

func (pa *ProjectAnalyzer) configureJava(dc *DockerConfig) {
	if pa.hasFile("pom.xml") {
		pa.configureMavenJava(dc)
	} else {
		pa.configureGradleJava(dc)
	}
}

func (pa *ProjectAnalyzer) configureMavenJava(dc *DockerConfig) {
	dc.baseImage = "eclipse-temurin:17-jdk-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY pom.xml .",
		"COPY .mvn .mvn",
		"COPY mvnw .",
		"RUN chmod +x mvnw",
		"RUN ./mvnw dependency:go-offline",
		"COPY src src",
		"RUN ./mvnw package -DskipTests",
		"EXPOSE 8080",
		"CMD [\"java\", \"-jar\", \"target/*.jar\"]",
	}
	dc.ports = []string{"8080:8080"}
}

func (pa *ProjectAnalyzer) configureGradleJava(dc *DockerConfig) {
	dc.baseImage = "eclipse-temurin:17-jdk-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY build.gradle settings.gradle ./",
		"COPY gradle gradle",
		"COPY gradlew .",
		"RUN chmod +x gradlew",
		"RUN ./gradlew dependencies",
		"COPY src src",
		"RUN ./gradlew build -x test",
		"EXPOSE 8080",
		"CMD [\"java\", \"-jar\", \"build/libs/*.jar\"]",
	}
	dc.ports = []string{"8080:8080"}
}

func (pa *ProjectAnalyzer) configureRuby(dc *DockerConfig) {
	dc.baseImage = "ruby:3.2-alpine"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY Gemfile Gemfile.lock ./",
		"RUN apk add --no-cache build-base postgresql-dev",
		"RUN bundle install",
		"COPY . .",
		"EXPOSE 3000",
		"CMD [\"bundle\", \"exec\", \"rails\", \"server\", \"-b\", \"0.0.0.0\"]",
	}
	dc.ports = []string{"3000:3000"}
	dc.environment = map[string]string{
		"RAILS_ENV": "production",
	}
}

func (pa *ProjectAnalyzer) configurePHP(dc *DockerConfig) {
	dc.baseImage = "php:8.2-apache"
	dc.commands = []string{
		"WORKDIR /var/www/html",
		"RUN apt-get update && apt-get install -y \\\n" +
			"    libzip-dev \\\n" +
			"    zip \\\n" +
			"    && docker-php-ext-install zip pdo pdo_mysql",
		"COPY --from=composer:latest /usr/bin/composer /usr/bin/composer",
		"COPY composer.* ./",
		"RUN composer install --no-dev --no-scripts --no-autoloader",
		"COPY . .",
		"RUN composer dump-autoload --optimize",
		"RUN chown -R www-data:www-data /var/www/html",
		"EXPOSE 80",
	}
	dc.ports = []string{"80:80"}
}

func (pa *ProjectAnalyzer) configureGeneric(dc *DockerConfig) {
	dc.baseImage = "ubuntu:latest"
	dc.commands = []string{
		"WORKDIR /app",
		"COPY . .",
		"CMD [\"/bin/bash\"]",
	}
}
