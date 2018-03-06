---
title: "Develop"
menu: 
    main:
        weight: 3
subnav: "true"
---
# Develop

This page describes the development environment installation of lorhammer and gives some architecture explanations.

## Install

### Requirement

* [Go](https://golang.org/doc/install) >= 1.9
* [Docker](https://docs.docker.com/engine/installation/) & [Docker-compose](https://docs.docker.com/compose/install/).
* [make command](http://www.tutorialspoint.com/unix_commands/make.htm)

### Steps

```shell
cd $GOPATH/src
git clone git@gitlab.com:itk.fr/lorhammer.git
cd lorhammer
make
```

The binaries of lorhammers are created in `./build` directory.

### First start

Follow the [quickstart](../quickstart) and be sure to have lorhammer, orchestrator and tools working.

## Documentation

### Engine

We use [hugo](https://gohugo.io/) to generate static html from markdown.
You can find documentation files in multiple directory.
All `.md` files at root path (README, CHANGELOG...) are used. We also use `doc/content/*.md`.
The theme can be find in `doc/themes/hugorha` after the first call to the `make doc` (see below).

### Generate doc

```shell
make doc
```

This script will install all requirements and generate the doc.

### Develop doc

```shell
make doc-dev
```

To launch a standalone web browser add `-dev` flag and open [http://localhost:1313/](http://localhost:1313/).
Each time you modify a doc file, the doc will be refresh in your browser.

## Architecture points

### Add a test type

A test type is a launcher of gateways. It describes how the orchestrator will build gateways to trigger the scenario.

Today we have 3 types of tests, the 'none' type does nothing, the 'one shot' type builds and launches nbGateway, the 'repeat' type is the same as 'one shot' but creates nbGateway of gateways every repeatTime.

A test type is a function that takes a Test, a model.Init and a client tool.mqtt. This function is started in a go routine and must call command.LaunchScenario(mqttClient, model.init) when nbGateway need to be launched.

The 'none' implementation is the most simple :

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

### Add a provisioner

A provisioner permit to register sensors and gateways to a network-server. Like that the network-server can accept messages from sensors and gateways.

Today we have 3 kind of provisioner : none to not provision, loraserver to provision a [loraserver](https://docs.loraserver.io) network server and a generic HTTP provisioner.

The HTTP provisioner send an http post to the *creationApiUrl*, its body is the [model godoc](/godoc/#model) in JSON format.

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
    provisioners[HttpType] = NewHttpProvisioner
}
```

Test it by creating a scenario file with your provisioning type and the configuration required by it.

We will happy to see lot of implementations of provisioner for different network-server open-source or proprietary.

### Add a deployer

A deployer permit to the orchestrator to deploy and instantiate lorhammers.

Today we have 5 kind of deployer. None to do nothing. Local to launch a local (sub-process) instance of lorhammer. Distant to scp and start over ssh lorhammers on other server. And amazon to provision server on [aws](https://aws.amazon.com/), deploy and launch lorhammers.

To add a deployer you need to have a fabric function which take config json.RawMessage and a mqtt client in parameters and return an implementation of `Deployer` interface :

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

func NewNone(_ json.RawMessage, _ tools.Mqtt) (Deployer, error) {
    return none{}, nil
}
```

A deployer need to have a type :

```go
const TypeNone = Type("none")
```

And you need to register your implementation in the map hosted by `src/orchestrator/deploy/deploy.go` :

```go
var deployers = make(map[Type]func(config json.RawMessage, mqttClient tools.Mqtt) (Deployer, error))

func init() {
    deployers[TypeNone] = NewNone
    deployers[TypeDistant] = NewDistantFromJson
    deployers[TypeAmazon] = NewAmazonFromJson
    deployers[TypeLocal] = NewLocalFromJson
}
```

We can image adding new deployer like [DigitalOcean](https://www.digitalocean.com/), [Kubernetes](https://kubernetes.io/) or [Swarm](https://docs.docker.com/engine/swarm/)...

### Add a checker

A checker is a way to verify that what have been sent by lorhammers have been correctly received by the IOT plateform. It aims to count the messages sent by lorhammers and check if result is equal to the expected value.
Checkers also allow to test if platform has accepted x messages (and not only received them). Useful for continious integration, orchestrator will exit(1) if all checks don't pass.

Today we have 3 kinds of checkers : 'none' for no checker, 'prometheus' to check number of messages sent againt the ones recieved by lorhammers and 'kafka' to check the content of messages.

To add a checker you need to have a factory function that takes a config json.RawMessage as parameter and returns an implementation of `Checker` interface :

```go
type CheckerSuccess interface {
    Details() map[string]interface{}
}

type CheckerError interface {
    Details() map[string]interface{}
}

type Checker interface {
    Check() ([]CheckerSuccess, []CheckerError)
}
```

The 'none' implementation example :

```go
type none struct{}

func newNone(_ json.RawMessage) (Checker, error) {
    return none{}, nil
}

func (_ none) Check() ([]CheckerSuccess, []CheckerError) {
    return make([]CheckerSuccess, 0), make([]CheckerError, 0)
}
```

A checker needs to have a type :

```go
const NoneType = Type("none")
```

A registration is needed for your implementation in the map hosted by `src/orchestrator/checker/checker.go` :

```go
var checkers = make(map[Type]func(config json.RawMessage) (Checker, error))

func init() {
    checkers[NoneType] = newNone
    checkers[PrometheusType] = newPrometheus
    checkers[kafkaType] = newKafka
}
```

We can imagine adding a file checker, if your application outputs processing on csv format, you will be abble to check if the csv is confirmed with the expected results. Execute it every time you push new code on your platform what enables you to have a continuous integration testing suite.