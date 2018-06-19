---
title: "Quickstart"
menu: 
    main:
        weight: 2
subnav: "true"
---
# Very Quickstart

For details let's go to next section.

## build (optional)

```bash
make

```

## start mandatory tools

```shell
./resources/scripts/launchTools.sh
```

## start 1 lorhammer worker

```shell
lorhammer -mqtt tcp://ADVERTISED_HOST:1884
```

## launch a scenario

```shell
orchestrator -mqtt tcp://ADVERTISED_HOST:1884 -from-file "./resources/scenarios/simple.json"
```

# Quickstart

This page describes how to run lorhammer from the simplest to more complex use case.

# Download

To download the binaries, go on [tags](https://gitlab.com/itk.fr/lorhammer/tags) page and click on the "download icon" next to the last release. Then select the `Download 'compile:tag'` item in the `Artifacts` category.

[![download-image](/images/quickstart/download.png)](/images/quickstart/download.png)

Inside the downloaded zip you will find lorhammer binaries for linux, mac and windows.

# Simple launch

To start only one lorhammer simulating 10 gateways with 5 nodes per gateway :

```shell
lorhammer -nb-gateway 10 -min-nb-node 5 -max-nb-node 5 -ns-address 127.0.0.1:1700
```

Lorhammer will start to stress the network server at `127.0.0.1:1700`. Except some logs, you will not see the details of what happens on your system. Follow next step for a more detailed run.

# Launch tools

Git clone the project (you don't need to have go installed to launch the tools).

To start the tools you need to have [docker](https://docs.docker.com/engine/installation/) and [docker-compose](https://docs.docker.com/compose/install/).

## Command

Start :

```shell
./resources/scripts/launchTools.sh
```

Will launch :

* [Mosquitto](https://mosquitto.org/) mqtt-broker to enable communication between orchestrator and lorhammer

# Launch orchestrator

One orchestrator can manage as much lorhammers as you want.

To start some lorhammers, launch the binary as shown below:

```shell
lorhammer -mqtt tcp://127.0.0.1:1884
lorhammer -mqtt tcp://127.0.0.1:1884
lorhammer -mqtt tcp://127.0.0.1:1884
```

Start an orchestrator with a simple scenario :

```shell
orchestrator -mqtt tcp://127.0.0.1:1884 -from-file "./resources/scenarios/simple.json"
```

This scenario will incrementally launch 10 gateways (going from 0 to 10 in 5 minutes). Each gateway will have 50 nodes. After 10 minutes `orchestrator` will check some numbers in prometheus and exit 1 if some check fails.

# Full example

A working example is accessible in `resources/examples/loraserver`. After cloning lorhammer go inside and read the [https://gitlab.com/itk.fr/lorhammer/tree/master/resources/examples/loraserver/README.md](README.md).

# First scenario

A scenario is an array of tests. A test is the description needed by the orchestrator (and the lorhammers) to stress a network server. All parameters are :

```json
[{
  "test": {
    "type": "none | oneShot | repeat",
    "repeatTime": "0"
  },
  "stopAllLorhammerTime": "0",
  "sleepBeforeCheckTime": "10s",
  "shutdownAllLorhammerTime": "10m",
  "sleepAtEndTime": "0",
  "init": {
    "nsAddress": "127.0.0.1:1700",
    "nbGatewayPerLorhammer": 10,
    "nbNodePerGateway": [50, 50],
    "nbScenarioReplayLaps": 0,
    "scenarioSleepTime": ["10s", "10s"],
    "receiveTimeoutTime": "1s",
    "gatewaySleepTime": ["100ms", "500ms"],
    "randomPayloads": false,
    "payloads" : [
      {"value": "01B501002919000006018403131313121233", "date": 1488931200},
      {"value": "01B501002919000006018403131313121244", "date": 1488931201}
    ]
  },
  "provisioning": {
    "type": "none | loraserver | http",
    "config": {
      "apiUrl": "127.0.0.1:9999",
      "abp": true,
      "login": "admin",
      "password": "admin",
      "organizationId": "",
      "networkServerId": "",
      "networkServerAddr": "",
      "serviceProfileID": "",
      "appId": "",
      "deleteOrganization": false,
      "deleteApplication": false,
      "nbProvisionerParallel": 10
    },
    "config": {
      "creationApiUrl": "http://127.0.0.1:1080",
      "deletionApiUrl": "http://127.0.0.1:1080"
    }
  },
  "check": {
    "type": "none | prometheus | kafka | mqtt",
    "config": {
      "address": "http://127.0.0.1:9090",
      "checks": [
        {"query": "sum(lorhammer_gateway)", "resultMin": 10, "resultMax": 10, "description": "nb gateways"}
      ]
    },
    "config": {
      "address": ["127.0.0.1:9092"],
      "topic": "test",
      "checks": [
        {
          "description": "MY_DATA",
          "remove": ["\"time\":[^,]+,","\"applicationID\":[^,]+,"],
          "text": "DATA"
        }
      ]
    },
    "config": {
      "address": "127.0.0.1:1883",
      "channel": "test",
      "checks": [
        {
          "description": "MY_DATA",
          "remove": ["\"time\":[^,]+,","\"applicationID\":[^,]+,"],
          "text": "DATA"
        }
      ]
    }
  },
  "deploy": {
    "type": "local | distant | amazon",
    "config": {
      "pathFile": "./build/lorhammer",
      "cleanPreviousInstances": true,
      "nbInstanceToLaunch": 1,
      "port": 1234
    },
    "config": { 
      "instances": [
        {
          "sshKeyPath": "",
          "user": "",
          "ipServer": "",
          "pathFile": "",
          "pathWhereScp": "",
          "beforeCmd": "",
          "afterCmd": "",
          "nbDistantToLaunch": 0
        }
      ]
    }
    },
    "config": {
      "region": "eu-west-2",
      "imageId": "ami-87848ee3",
      "instanceType": "t2.micro",
      "keyPairName": "amazon-pc_itk_romain",
      "securityGroupIds": ["sg-9372c1fa"],
      "minCount": 10,
      "maxCount": 10,
      "distantConfig": {
        "sshKeyPath": "~/.ssh/amazon-pc_itk_romain",
        "user": "admin",
        "pathFile": "./build/lorhammer",
        "pathWhereScp": "/home/admin/",
        "nbDistantToLaunch": 1
      }
    }
  }
}]
```

# Scenario Parameters

## test

Type : **object/struct**

On object to describe test parameters

### type

Type : **string/enum**

Can be `none`, `oneShot` or `repeat`

* `none` no test will be launched, useful to use an orchestrator to deploy lorhammer instances for a future use
* `oneShot` starts init.nbGatewayPerLorhammer with init.nbNodePerGateway[0] >= nbNode <= init.nbNodePerGateway[1]
* `repeat` starts init.nbGatewayPerLorhammer every `repeatTime`

### repeatTime

Type : **string/duration**

If `testType` == `repeat` used to create init.nbGatewayPerLorhammer every time

## stopAllLorhammerTime

Type : **string/duration**

When this period is over, the  orchestrator stops all scenarios running on lorhammers (note that the lorhammer instances are still running in this case), 0 if you want the scenarios to run continuously

## sleepBeforeCheckTime

Type : **string/duration**

When this period is over, the  orchestrator will check if results are good defined in the check property.

## shutdownAllLorhammerTime

Type : **string/duration**

When this period is over, the orchestrator shuts down all lorhammer instances, 0 for all instances to be running for an undetermined period of time

## sleepAtEndTime

Type : **string/duration**

After the current test (including time to stress, stop and shutdown), this duration will be used to make the orchestrator sleep. Could be Useful to make a pause between tests.

## init

Type : **object/struct**

Represents the lorawan protocol

### nsAddress

Type : **string/address**

The ip:port of network-server to stress

### nbGatewayPerLorhammer

Type : **int** : The number of gateways to create per lorhammer

### nbNodePerGateway

Type : **int,int**

The minimum and maximum number of nodes to instantiate per gateway. A random number between the given range will be used. Use the same value for min,max not to randomize.

### nbScenarioReplayLaps

Type : **int** : The number of times the entire set of payloads will be sent. "Zero" means infinite, all the scenario will be stopped regarding the temporal settings.

### scenarioSleepTime

Type : **string/duration, string/duration**

This represents the time interval between every data sent from all simulated gateways to the network server, an array value allow you to randomize (min, max)

### receiveTimeoutTime

Type : **string/duration**

This represents the timeout period for Lora Server ack time.

### gatewaySleepTime

Type : **string/duration, string/duration**

This represents the time interval between every data sent of each gateway to network server, an array value allow you to randomize (min, max)

### appskey

Type : **optional(string)**

This parameter should be present when using an activation by personalization (see: **abp**) with the application server.

### nwskey

Type : **optional(string)**

This parameter should be present when using an activation by personalization (see: **abp**) with the application server. This key is used to encrypt all push data payloads

### payloads

Type : **array(model.Payload)**
> For more details read the [godoc](/godoc/#type-testsuite)

This array holds the different payloads you want the nodes to send through all their messages.
Each node will send one of the payloads at a time (in the array natural order or randomly if the `randomPayloads` property is set to true).
The payloads here are hexadecimal string representations.
A date property can be added for each payload, and will be used to set the rxpkDate of the frame.
This can be helpful if that date is used as an absolute time reference to timestamp the measures.

### randomPayloads

Type : **boolean**

If 'true', take randomly content from payload array. If 'false' take successivly content from payload array

### withJoin

Type : **boolean**

This is used when provisioning is "true" active. If 'true', all nodes will join the network with a join request. (TODO : For now, the JoinAccept message still need to be processed. )

### rxpkDate

Type : **long** epoch time in seconds

This is used when for the dates sent with rxpk messages. It's useful when we want to have control over dates, especially when using a specific checker to validate a dates...

### Description

Type : **optional(string)**

Description sended with node provisioning

## provisioning

Type : **object/struct**

Describes the provisioning of your sensors on the network-server system

### type

Type : **string/enum** : Can be `none`, `loraserver` or `http`

* `none` no provisioning is required
* `loraserver` call the api of [loraserver application server](https://docs.loraserver.io/loraserver/application-server/) and add sensors
* `http` send HTTP post, have a look at [model godoc](/godoc/#model) for body JSON format

### config

Type : **object/struct**

Depending on the provisioner you chose, these are the optional fields :

#### apiUrl

Type : **optional(string)**

Api url for lorawanserver.

#### abp

Type : **optional(boolean)**

When this flag is at 'true', the activation by personalization is activated for the provisioning. Note that in this case, the **appskey** and **nwskey** are mandatory on the 'init' descriptor

#### login

Type : **optional(string)**

The provided application server username.

#### Password

Type : **optional(string)**

The provided application server password.

#### organizationId

Type : **optional(string)**

if empty create new `lorhammer` organization in loraserver or use define organizationId.

#### networkServerId

Type : **optional(string)**

if empty create new networkServerId in loraserver or use define networkServerId.

#### networkServerAddr

Type : **optional(string)**

if empty you need to have set `networkServerId`, lorhammer create the networkServer with this networkServerAddr if not already created.

#### serviceProfileID

Type : **optional(string)**

if empty create new serviceProfileID in loraserver or use define serviceProfileID.

#### appId

Type : **optional(string)**

if empty create new app in loraserver or use define appId.

#### deleteOrganization

Type : **bool**

Delete organization at the end of test.

#### deleteApplication

Type : **bool**

Delete application at the end of test.

#### nbProvisionerParallel

Type : **int**

Number of parallel request will access loraserver to provision.

#### creationApiUrl

Type : **optional(string)**

API URL to call by HTTP provisioner for creation

#### deletionApiUrl

Type : **optional(string)**

API URL to call by HTTP provisioner for deletion

## check

Type : **object/struct**

Describes the check orchestrator must do at the end of scenario

### type

Type : **string/enum** : Can be `none`, `prometheus`, `kafka` ot `mqtt`

* `none` no check is required
* `prometheus` call the api of prometheus and compare results
* `kafka` listen kafka queue and check if messages are good
* `mqtt` listen mqtt brocker and check if messages are good

### prometheus config

Type : **object/struct**

Allows to check, at the end of a test, if the results are good or not depending on what you want, useful for ci check (exit 1 if check fail)

* address **string** : the prometheus address scheme://ip:port
* checks **array(object)** : A logic `and` will be executed on each checks described bellow
  * query **string** : the prometheus query to execute
  * resultMin **int** : the result min you want (put min == max when you expect an exact value )
  * resultMax **int** : the result max you want (put min == max when you expect an exact value)
  * description **string** : the description logged if check fail

### kafka config

Type : **object/struct**

* address **array(string)** : the kafka ip:port of brokers
* topic **string** : the kafka topic to listen
* checks **array(object)** : A logic `or` will be executed on each checks described bellow
  * description **string** : the description logged if check fail
  * remove **array(string)** : An array of regexp to clean random/dynamic data produced by the test (timestamp...)
  * text **string** : The text to check

### mqtt config

Type : **object/struct**

* address **string** : the mqtt ip:port of broker
* channel **string** : the mqtt channel to listen
* checks **array(object)** : A logic `or` will be executed on each checks described bellow
  * description **string** : the description logged if check fail
  * remove **array(string)** : An array of regexp to clean random/dynamic data produced by the test (timestamp...)
  * text **string** : The text to check

## deploy

Type : **object/struct**

Allows to deploy a lorhammer before launching the tests

### type

Type : **string/enum**

Can be `none`, `local`, `distant` or `amazon`

* `none` no deployment is made
* `local` runs a sub-process with lorhammer on the same mqtt that has the current orchestrator
* `distant` performs an scp to send `deploy.config.pathFile` to a distant server and runs ssh to start it
* `amazon` uses amazon api to create aws instances and run lorhammers on the go

### config

> For more details read the [godoc](/godoc/#type-testsuite)

# Tips

## All flags

Display all flags available in lorhammer or in orchestrator :

```shell
lorhammer -help
orchestrator -help
```

## Log tools

To see logs of tools, useful to debug, at the root of lorhammer enter :

```shell
docker-compose logs
```

## Orchestrator cli

You can launch orchestrator in cli mode to have some utilities (stop current scenarios, shutdown lorhammers...)

```shell
ochestrator -mqtt tcp://127.0.0.1:1883 -cli
```
