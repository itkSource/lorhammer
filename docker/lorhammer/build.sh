#!/bin/bash

set -o errexit
set -o pipefail

readonly BASEDIR=$(dirname $(readlink -f $0))

docker build -t "registry.gitlab.com/itk.fr/lorhammer/lorhammer" -f ${BASEDIR}/Dockerfile-lorhammer ${BASEDIR}/../..
docker build -t "registry.gitlab.com/itk.fr/lorhammer/orchestrator" -f ${BASEDIR}/Dockerfile-orchestrator ${BASEDIR}/../..
