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
git clone git@gitlab.com:itk.fr/lorhammer.git
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

A test type is a launcher of gateways. It's describe how the orchestrator will build gateways to realize the scenario.

Today we have 4 types of test, none do nothing, one shot build nbGateway and launch them, repeat is the same as one shot but repeat creation of nbGateway every repeatTime and ramp do the same but distribute the creation of gateway to have nbGateway after the rampTime.

A test type is a function which take a Test, a model.Init and a client tool.mqtt. This function will be started in a go routine and must call command.LaunchScenario(mqttClient, model.init) when nbGateway need to be launched.

The none implementation is the most simple :

```go
func startNone(_ Test, _ model.Init, _ tools.Mqtt) {
	LOG_NONE.WithField("type", "none").Warn("Nothing to test")
}
```

You can find other implementations in `src/orchestrator/testType` package.

A test type is also represented by a testType.Type type like that :

```go
const TypeNone Type = "none"
```

To finish you need to add your implementation in the map hosted by `src/orchestrator/testType/testType.go` :

```go
var testers = make(map[Type]func(test Test, init model.Init, mqttClient tools.Mqtt))

func init() {
	testers[TypeNone] = startNone
}
```

Test it by creating a scenario file with your test type.

We can imagine make `pic` test with creation and deletion of gateways over the time. Or a ramp node test which build node in existing gateway over the time.

## Add a provisioner

A provisioner permit to register sensors and gateways to a network-server. Like that the network-server can accept messages from sensors and gateways.

Today we have 3 kind of provisioner : none to not provision, loraserver to provision a loraserver network server and semtechv4 to provision a semtech v4 network server.
 
The semtechv4 provisioner is a work in progress, any help to do it will be useful. Please add comments in [issues/13](https://gitlab.com/itk.fr/lorhammer/issues/13) if you want to contribute on it.
 
A provisioner must have a fabric function which take config json.RawMessage in parameters and return an implementation of `Provisioner` interface :

```go
type provisioner interface {
	Provision(sensorsToRegister model.Register) error
	DeProvision() error
}
```

The simpler implementation is the NoneProvisioner :

```go
type none struct{}

func NewNone(_ json.RawMessage) (provisioner, error) { return none{}, nil }

func (_ none) Provision(sensorsToRegister model.Register) error { return nil }

func (_ none) DeProvision() error { return nil }
```

The fabric function must have a provisioning.Type :

```go
const NoneType = Type("none")
```

And you need to register your implementation in the map hosted by `src/orchestrator/provisioning/provisioning.go` :

```go
var provisioners = make(map[Type]func(config json.RawMessage) (provisioner, error))

func init() {
	provisioners[NoneType] = NewNone
	provisioners[LoraserverType] = NewLoraserver
	provisioners[SemtechV4Type] = NewSemtechV4
}
```

Test it by creating a scenario file with your provisioning type and the configuration required by it.

We will happy to see lot of implementations of provisioner for different network-server open-source or proprietary. 

## Add a deployer

A deployer permit to the orchestrator to deploy and instantiate lorhammers. 

Today we have 5 kind of deployer. None to do nothing. Local to launch a local (sub-process) instance of lorhammer. Distant to scp and start lorhammers on other server. And Amazone to provision server on amazon, deploy and launch lorhammers. 

To add a deployer you need to have a fabric function which take config json.RawMessage and a consul client in parameters and return an implementation of `Deployer` interface :

```go
type Deployer interface {
	RunBefore() error
	Deploy() error
	RunAfter() error
}
```

The None implementation is really simple :

```go
type none struct{}

func (_ none) RunBefore() error { return nil }
func (_ none) Deploy() error    { return nil }
func (_ none) RunAfter() error  { return nil }

func NewNone(_ json.RawMessage, _ tools.Consul) (Deployer, error) {
	return none{}, nil
}
```

A deployer need to have a type :

```go
const TypeNone = Type("none")
```

And you need to register your implementation in the map hosted by `src/orchestrator/deploy/deploy.go` :

```go
var deployers = make(map[Type]func(config json.RawMessage, consulClient tools.Consul) (Deployer, error))

func init() {
	deployers[TypeNone] = NewNone
	deployers[TypeDistant] = NewDistantFromJson
	deployers[TypeAmazon] = NewAmazonFromJson
	deployers[TypeLocal] = NewLocalFromJson
}
```

We can image adding new deployer like DigitalOcean, Kubernetes or Swarm...
