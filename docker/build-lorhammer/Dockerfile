FROM golang:1.10.1

# Used to build lorhammer project, from golang and add docker-compose (docker is a service in gitlab-ci.yml)
# Push a new version :
# > docker login registry.gitlab.com
# > docker build -t registry.gitlab.com/itk.fr/lorhammer/build .
# > docker push registry.gitlab.com/itk.fr/lorhammer/build

##
# Docker
##
RUN set -x \
    && echo 'Acquire::http::Pipeline-Depth "0";' > /etc/apt/apt.conf.d/http-pipeline \
	&& curl -fsSL get.docker.com -o get-docker.sh \
	&& sh get-docker.sh \
	&& docker -v

##
# Docker compose
##
ENV DOCKER_COMPOSE_VERSION 1.20.1
RUN curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose \
    && docker-compose -v

##
# Golang dep
##
RUN curl -L "https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64" -o ${GOPATH}/bin/dep \
	&& chmod +x ${GOPATH}/bin/dep \
	&& dep version

##
# Doc hugo
##
RUN curl -L "https://github.com/gohugoio/hugo/releases/download/v0.38.1/hugo_0.38.1_Linux-64bit.tar.gz" -o /tmp/hugo.tar.gz \
	&& mkdir /tmp/hugo && tar xf /tmp/hugo.tar.gz -C /tmp/hugo \
	&& mv /tmp/hugo/hugo /usr/local/bin/hugo \
    && rm /tmp/hugo.tar.gz && rm -rf /tmp/hugo \
    && chmod +x /usr/local/bin/hugo \
	&& hugo version

##
# Doc others
##
RUN go get github.com/robertkrimen/godocdown/godocdown \
	&& go get github.com/tdewolff/minify/cmd/minify
