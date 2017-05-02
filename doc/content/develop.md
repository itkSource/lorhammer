---
title: "Develop"
menu: 
    main:
        weight: 3
subnav: "true"
---

# Develop

This page describes the development environment installation of lorhammer and gives some architecture explanations.

# Install

## Requirement

* [Go](https://golang.org/doc/install) >= 1.8
* [Docker](https://docs.docker.com/engine/installation/) & [Docker-compose](https://docs.docker.com/compose/install/).

## Steps

```shell
cd $GOPATH/src
git clone git@gitlab.lan.itkweb.fr:platform-iot/lorhammer.git
cd lorhammer
sh install.sh
```

## Compilation

```shell
sh build.sh
```

The binaries of lorhammers are created in `./build` directory.

## First start

Follow the [quickstart](quickstart) and be sure to have lorhammer, orchestrator and tools working.

# Documentation

## Engine

We use [hugo](https://gohugo.io/) to generate static html from markdown.
You can find documentation files in multiple directory. 
All `.md` files at root path (README, CHANGELOG...) are used. We also use `doc/content/*.md`.
The theme can be find in `doc/themes/hugorha` after the first call to the `makeDoc.sh` (see below).

## Generate doc

```shell
./resources/scripts/makeDoc.sh
```

This script will install all requirements and generate the doc. 

## Develop doc

```shell
./resources/scripts/makeDoc.sh -dev
```

To launch a standalone web browser add `-dev` flag and open [http://localhost:1313/](http://localhost:1313/).
Each time you modify a doc file, the doc will be refresh in your browser.

# Architecture points

## Add a test type

TODO make it

## Add a provisioner

## Add a deployer

## Add personal push data