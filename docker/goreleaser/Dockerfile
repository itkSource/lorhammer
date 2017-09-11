FROM golang:1.9-alpine

# Used to build lorhammer project with goreleaser
# Push a new version :
# > docker login registry.gitlab.com
# > docker build -t registry.gitlab.com/itk.fr/lorhammer/goreleaser .
# > docker push registry.gitlab.com/itk.fr/lorhammer/goreleaser

RUN apk update --no-cache \
   && apk add --no-cache ca-certificates git wget ruby ruby-dev build-base libffi-dev tar rpm \
   && update-ca-certificates &>/dev/null \
   && wget -q -O /tmp/goreleaser.tar.gz \
      https://github.com/goreleaser/goreleaser/releases/download/v0.30.4/goreleaser_Linux_x86_64.tar.gz \
   && mkdir /tmp/goreleaser && tar xf /tmp/goreleaser.tar.gz -C /tmp/goreleaser \
   && rm /tmp/goreleaser.tar.gz \
   && gem install --no-ri --no-rdoc fpm

WORKDIR /go/src/lorhammer/

ENTRYPOINT ["/tmp/goreleaser/goreleaser", "--skip-publish"]