#!/usr/bin/env bash



installDepthAndLaunch() {
    if hash hugo 2>/dev/null; then
        echo "hugo found no need to install"
    else
        go get -u -v github.com/gohugoio/hugo
    fi

    if hash godocdown 2>/dev/null; then
        echo "godocdown found no need to install"
    else
        go get github.com/robertkrimen/godocdown/godocdown
    fi

    if hash minify 2>/dev/null; then
        echo "minify found no need to install"
    else
        go get github.com/tdewolff/minify/cmd/minify
    fi

    goDoc

    if [[ "$1" == "-dev" ]]; then
        env HUGO_BASEURL="http://127.0.0.1:1313/" hugo server -wDs doc
    else
        hugo -s doc
        minify --recursive --output ./doc/public_min/ ./doc/public
        cp -u -r ./doc/public/images/. ./doc/public_min/images
        cp -u -r ./doc/public/font/. ./doc/public_min/font
    fi
}

goDoc() {
    echo -e "---\ntitle: \"GoDoc\"\nmenu: \n    main:\n        weight: 21\nsubnav: \"true\"\n---\n" > doc/content/godoc.md
    for p in `go list ./src/...`
    do
        godocdown ${p} >> doc/content/godoc.md
    done
}

###
# $1 file will be copied
# $2 file where copy
# $3 title of generated page
# $4 weight of the menu (bigger is right)
# $5 subnav "true" or "false"
###
addFile() {
    cp $1 $2
    sed -i -E "s/doc\/static\/images\/([^\)]+)/\/images\/\1/g" $2
    echo -e "---\ntitle: \"$3\"\nmenu: \n    main:\n        weight: $4\nsubnav: \"$5\"\n---\n$(cat $2)" > $2
}

if [ ! -d "doc/themes/hugorha" ]; then
    mkdir -p doc/themes/hugorha
    git clone https://github.com/itkSource/hugorha.git doc/themes/hugorha
fi

clean() {
    rm doc/content/_index.md
    rm doc/content/CHANGELOG.md
    rm doc/content/CONTRIBUTING.md
    rm doc/content/LICENCE.md
    rm doc/content/godoc.md
}

addFile README.md doc/content/_index.md "Lorhammer" 1 "true"
addFile CHANGELOG.md doc/content/CHANGELOG.md "Changelog" 10 "true"
addFile CONTRIBUTING.md doc/content/CONTRIBUTING.md "Contributing" 20 "true"
addFile LICENCE.md doc/content/LICENCE.md "Licence" 30 "false"

trap 'clean' 2 3

installDepthAndLaunch $1

if [[ "$1" != "-dev" ]]; then
    clean
fi
