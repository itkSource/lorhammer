FROM golang:1.8-alpine

RUN apk update --no-cache \
   && apk add --no-cache ca-certificates git wget ruby ruby-dev build-base libffi-dev tar rpm \
   && update-ca-certificates &>/dev/null \
   && wget -q -O /tmp/goreleaser.tar.gz \
      https://github.com/goreleaser/goreleaser/releases/download/v0.27.4/goreleaser_Linux_x86_64.tar.gz \
   && mkdir /tmp/goreleaser && tar xf /tmp/goreleaser.tar.gz -C /tmp/goreleaser \
   && gem install --no-ri --no-rdoc fpm

ADD ./.goreleaser.yml ./README.md ./LICENCE.md ./CHANGELOG.md ./CONTRIBUTING.md /go/src/lorhammer/
ADD ./.git /go/src/lorhammer/.git
ADD ./src /go/src/lorhammer/src
ADD ./vendor /go/src/lorhammer/vendor

WORKDIR /go/src/lorhammer/

ENTRYPOINT ["/tmp/goreleaser/goreleaser", "--skip-publish"]