#!/bin/bash

name=$(date +%s)
path="/tmp/coverage-lorhammer-$name.cov"

generate() {
    set -e
    echo 'mode: count' > "$path"
    for dir in $(find ./src -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d);
    do
    if ls $dir/*.go &> /dev/null; then
        go test -short -covermode=count -coverprofile=$dir/profile.tmp $dir
        if [ -f $dir/profile.tmp ]
        then
            cat $dir/profile.tmp | tail -n +2 >> "$path"
            rm $dir/profile.tmp
        fi
    fi
    done
}

displayTerminal() {
    go tool cover -func "$path"
}

displayHtml() {
    go tool cover -html "$path"
}

usage () {
    cat << EOF

Description: Go coverage for all packages.

Usage: resources/scripts/cover.sh [COMMAND]

Commands:

-t | -terminal			Prompt the list of go packages with percent coverage.
-m | -html			Open default web browser with percent coverage.
-h | -help			Display this help

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

generate
if [[ "$1" == "-terminal" || "$1" == "-t" ]]; then
    displayTerminal; else
    displayHtml
fi