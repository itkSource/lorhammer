FROM golang:1.8

# Used to build lorhammer project, from golang and add docker-compose (docker is a service in gitlab-ci.yml)
# Push a new version :
# > docker login registry.gitlab.com
# > docker build -t registry.gitlab.com/itk.fr/lorhammer/build .
# > docker push registry.gitlab.com/itk.fr/lorhammer/build

# Docker compose
ENV DOCKER_COMPOSE_VERSION 1.12.0

RUN curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose
