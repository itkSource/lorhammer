#!/bin/bash

set -o errexit
set -o pipefail

readonly BASEDIR=$(dirname $(readlink -f $0))
CI_JOB_ID=$(printenv CI_JOB_ID)

docker build -t "lorhammer_goreleaser_$CI_JOB_ID" -f ${BASEDIR}/../../docker/goreleaser/Dockerfile ${BASEDIR}/../..

docker run --rm -v ${BASEDIR}/../../dist:/go/src/lorhammer/dist "lorhammer_goreleaser_$CI_JOB_ID" $@