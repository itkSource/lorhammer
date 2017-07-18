#!/bin/bash

set -o errexit
set -o pipefail

readonly BASEDIR=$(dirname $(readlink -f $0))
CI_JOB_ID=$(printenv CI_JOB_ID)
if [ -z "$CI_JOB_ID" ]; then
    CI_JOB_ID="0.0.0"
fi

docker build -t "lorhammer_goreleaser_$CI_JOB_ID" -f ${BASEDIR}/../../docker/goreleaser/Dockerfile ${BASEDIR}/../..

docker run --rm "lorhammer_goreleaser_$CI_JOB_ID" $@