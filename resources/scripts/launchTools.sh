#!/bin/bash

set -o errexit
set -o pipefail

ARGS=()
if [ ! -z "$CI_JOB_ID" ]; then
    ARGS+=( '-f' 'docker-compose.yml' '-f' 'docker-compose.integration.yml' '-p' "${CI_JOB_ID}" )
fi

echo "Start launchTools.sh"
echo "Clean docker deamon : docker-compose ${ARGS[@]} down --remove-orphans"
docker-compose ${ARGS[@]} down --remove-orphans
echo "Start tools : docker-compose ${ARGS[@]} up -d mqtt"
docker-compose ${ARGS[@]} up -d mqtt
