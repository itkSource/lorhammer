#!/bin/bash

set -o errexit
set -o pipefail

readonly BASEDIR=$(dirname $(readlink -f $0))

usage () {
    cat << EOF

Description: Build lorhammer binaries.

Usage: resources/scripts/buildAllEnv.sh [COMMAND]

Commands:

-light  		    Only compile linux amd64 version.
-full  		        Compile linux, window and mac for 386, amd64 and arm.
-h | -help			Display this help.

EOF

}

if [[ -z $1 ]]; then
    echo "Error : command empty"
    usage
    exit 1
fi

if [[ "$1" == "-help" || "$1" == "-h" ]]; then
    usage
    exit 0
fi

if [[ "$1" == "-light" ]]; then
    docker run --rm -v ${BASEDIR}/../..:/go/src/lorhammer registry.gitlab.com/itk.fr/lorhammer/goreleaser --config docker/goreleaser/goreleaser-light.yml "${@:2}";
    else
    docker run --rm -v ${BASEDIR}/../..:/go/src/lorhammer registry.gitlab.com/itk.fr/lorhammer/goreleaser --config docker/goreleaser/goreleaser-full.yml "${@:2}"
    find ${BASEDIR}/../../dist -maxdepth 1 -mindepth 1 -type d -exec rm -rf '{}' \;
fi