FROM golang:1.8

# Used to build lorhammer project, from golang and add docker-compose (docker is a service in gitlab-ci.yml)
# Push a new version :
# > docker login registry.gitlab.com
# > docker build -t registry.gitlab.com/itk.fr/lorhammer/build .
# > docker push registry.gitlab.com/itk.fr/lorhammer/build

##
# Versions
##
ENV DOCKER_VERSION 1.12.1
ENV DOCKER_COMPOSE_VERSION 1.12.0

##
# Docker
##
ENV DOCKER_BUCKET get.docker.com
ENV DOCKER_SHA256 05ceec7fd937e1416e5dce12b0b6e1c655907d349d52574319a1e875077ccb79

RUN set -x \
	&& curl -fSL "https://${DOCKER_BUCKET}/builds/Linux/x86_64/docker-${DOCKER_VERSION}.tgz" -o docker.tgz \
	&& echo "${DOCKER_SHA256} *docker.tgz" | sha256sum -c - \
	&& tar -xzvf docker.tgz \
	&& mv docker/* /usr/local/bin/ \
	&& rmdir docker \
	&& rm docker.tgz \
	&& docker -v

##
# Docker compose
##
RUN curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose

##
# Pygment for doc
##
RUN wget https://bootstrap.pypa.io/get-pip.py && python get-pip.py && pip install Pygments
