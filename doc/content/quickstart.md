---
title: "Quickstart"
menu: 
    main:
        weight: 2
subnav: "true"
---

# Quickstart

This page describes how to run lorhammer from the simplest to more complex use case.

# Simple launch

To start only one lorhammer simulating 10 gateways with 5 nodes per gateway :

```shell
lorhammer -nb-gateway 10 -min-nb-node 5 -max-nb-node 5 -ns-address 127.0.0.1:1700
```

Lorhammer will start to stress the network server at `127.0.0.1:1700`. Except some logs, you will not see the details of what happens on your system. Follow next step for a more detailed run.

# Launch tools

Git clone the project (you don't need to have go installed to launch the tools).

To start the tools you need to have [docker](https://docs.docker.com/engine/installation/) and [docker-compose](https://docs.docker.com/compose/install/).

## Environment variables

Add this environment variables :

* **LORHAMMER_CONSUL_IP** : the ip address used to contact consul
* **LORHAMMER_PROMETHEUS_IP** : the ip address used to run prometheus
* **LORHAMMER_MQTT_IP** : the ip address used to run mqtt used for cross tool communication
* **LORHAMMER_GRAFANA_IP** : the ip address used to access grafana

Most of the time, this variables will point at `127.0.0.1`.

> If you use the docker-compose file shipped with lorhammer you need to have **LORHAMMER_MQTT_PORT** = 1884 !

Alternatively you can override port variables (in case of double usage/installation, this ports must be free on host) :

* **LORHAMMER_MQTT_PORT** : the port address used for mqtt, default is 1883
* **LORHAMMER_PROMETHEUS_PORT** : the port used to communicate with prometheus, default is 9090
* **LORHAMMER_CONSUL_PORT** : the port used to communicate with consul, default is 8500
* **LORHAMMER_GRAFANA_PORT** : the port used to communicate with grafana, default is 3000

## Command

Start :

```shell
resources/scripts/launchTools.sh
```

Will launch :

* [Prometheus](https://prometheus.io/) to manage metrics ([web application](http://127.0.0.1:9090/))
* [Grafana](https://grafana.com/) to chart metrics from prometheus ([web application](http://127.0.0.1:3000/))
* [Consul](https://www.consul.io/) to discover lorhammer services
* [Mosquitto](https://mosquitto.org/) mqtt-broker to enable communication between orchestrator and lorhammer

[![lorhammer-schema](/images/Tools-schema.png)](/images/Tools-schema.png)

When all the tools are launched, open a web browser on [127.0.0.1:3000](127.0.0.1:3000). By default login is `admin` and password is `pass`. 
Add a data source with name `prometheus`, type `Prometheus`, url `lorhammer_prometheus_1:9090` and let other params with their default values.
Load default dashboard, you can find it here :  `resources/grafana/DashboardLora.json`.

Go to the `Lora` dashboard, if all is ok then start a lorhammer 

```shell
lorhammer -nb-gateway 10 -min-nb-node 5 -max-nb-node 5 -ns-address 127.0.0.1:1700 -consul 127.0.0.1:8500
```

You will see :

[![simple launch illustration](/images/quickstart/simpleLaunch.png)](/images/quickstart/simpleLaunch.png)

# Launch orchestrator

One orchestrator can manage as much lorhammers as you want.

To start some lorhammers, launch the binary as shown below:

```shell
lorhammer -consul 127.0.0.1:8500
lorhammer -consul 127.0.0.1:8500
lorhammer -consul 127.0.0.1:8500
```

Start an orchestrator with a simple scenario :

```shell
orchestrator -consul 127.0.0.1:8500 -from-file "./resources/scenarios/simple.json"
```

This scenario will incrementally launch 10 gateways (going from 0 to 10 in 5 minutes). Each gateway will have 50 nodes. After 10 minutes `orchestrator` will check some numbers in prometheus and exit 1 if some check fails.

Don't forget to open grafana dashboard to see what happens.

# First scenario

A scenario is an array of tests. A test is the description needed by the orchestrator (and the lorhammers) to stress a network server. All parameters are :

```json
[{
  "test": {
    "type": "oneShot | repeat | ramp",
    "rampTime": "5m",
    "repeatTime": "0"
  },
  "stopAllLorhammerTime": "0",
  "shutdownAllLorhammerTime": "10m",
  "init": {
    "nsAddress": "127.0.0.1:1700",
    "nbGatewayPerLorhammer": 10,
    "nbNodePerGateway": [50, 50],
    "scenarioSleepTime": ["10s", "10s"],
    "gatewaySleepTime": ["100ms", "500ms"]
  },
  "provisioning": {
    "type": "none",
    "config": {
      "apiUrl": "127.0.0.1:9999"
    },
    "config": {
      "nsAddress": "127.0.0.1:1701",
      "asAddress": "127.0.0.1:4000",
      "csAddress": "127.0.0.1:5000",
      "ncAddress": "127.0.0.1:6000"
    }
  },
  "prometheusCheck": [
    {"query": "sum(lorhammer_gateway)", "resultMin": 10, "resultMax": 10, "description": "nb gateways"}
  ],
  "deploy": {
    "type": "local | distant | amazon",
    "config": {
      "pathFile": "./build/lorhammer",
      "cleanPreviousInstances": true,
      "nbInstanceToLaunch": 1
    },
    "config": {
      "sshKeyPath": "",
      "user": "",
      "ipServer": "",
      "pathFile": "",
      "pathWhereScp": "",
      "beforeCmd": "",
      "afterCmd": "",
      "nbDistantToLaunch": 0
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

Can be `none`, `oneShot`, `repeat` or `ramp`

* `none` no test will be launched, useful to use an orchestrator to deploy lorhammer instances for a future use
* `oneShot` starts init.nbGatewayPerLorhammer with init.nbNodePerGateway[0] >= nbNode <= init.nbNodePerGateway[1]
* `repeat` starts init.nbGatewayPerLorhammer every `repeatTime`
* `ramp` starts init.nbGatewayPerLorhammer / rampTime every minute during `rampTime`
    
### rampTime
    
Type : **string/duration**
 
If `testType` == `ramp` then init.nbGatewayPerLorhammer reached incrementally throughout the duration of the scenario

### repeatTime 

Type : **string/duration**

If `testType` == `repeat` used to create init.nbGatewayPerLorhammer every time

## stopAllLorhammerTime 

Type : **string/duration**

When this period is over, the  orchestrator stops all scenarios running on lorhammers (note that the lorhammer instances are still running in this case), 0 if you want the scenarios to run continuously

## shutdownAllLorhammerTime 

Type : **string/duration**

When this period is over, the orchestrator shuts down all lorhammer instances, 0 for all instances to be running for an undetermined period of time

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

### scenarioSleepTime 

Type : **string/duration, string/duration**

This represents the time interval between every data sent from all simulated gateways to the network server, an array value allow you to randomize (min, max)

### gatewaySleepTime 

Type : **string/duration, string/duration**

This represents the time interval between every data sent of each gateway to network server, an array value allow you to randomize (min, max)

### appskey 

Type : **optional(string)**

This parameter should be present when using an activation by personalization (see: **isabp**) with the application server.

### nwskey 

Type : **optional(string)**

This parameter should be present when using an activation by personalization (see: **isabp**) with the application server. This key is used to encrypt all push data payloads

### payloads 

Type : **array(string)**

This array holds the different payloads you want the nodes to send through all their messages. Each node will randomly choose one of the payloads given in the array as the only payload he's going to be sending. The payloads here are hexadecimal string representations
 
### withJoin 
 
Type : **boolean**

This is used when provisioning is active. If 'true', all nodes will join the network with a join request. (TODO : For now, the JoinAccept message still need to be processed. )  


## provisioning 

Type : **object/struct**

Describes the provisioning of your sensors on the network-server system

### type 

Type : **string/enum** : Can be `none`, `brocaar` or `semtechv4`

* `none` no provisioning is required
* `brocaar` call the api of lorawanserver and add sensors
* `semtechv4` send tcp order to add gateways and sensors

### config

Type : **object/struct**

Depending on the provisioner you chose, these are the optional fields : 

#### apiUrl 

Type : **optional(string)**
 
Api url for lorawanserver.

#### isabp 

Type : **optional(boolean)**
 
When this flag is at 'true', the activation by personalization is activated for the provisioning. Note that in this case, the **appskey** and **nwskey** are mandatory on the 'init' descriptor

#### nsAddress 

Type : **optional(string)**
 
ip:port to reach semtechv4 network-server

#### asAddress 

Type : **optional(string)**

ip:port to reach semtechv4 application-server

#### csAddress 

Type : **optional(string)**

ip:port to reach semtechv4 customer-server

#### ncAddress 

Type : **optional(string)**

ip:port to contact semtechv4 netwrok-controller

## prometheusCheck 

Type : **array(object/struct)**

Allows to check, at the end of a test, if the results are good or not depending on what you want, useful for ci check (exit 1 if check fail)

* query **string** : the prometheus query to execute
* resultMin **int** : the result min you want (put min == max when you expect an exact value )
* resultMax **int** : the result max you want (put min == max when you expect an exact value)
* description **string** : the description logged if check fail
    
## deploy 

Type : **object/struct**

Allows to deploy a lorhammer before launching the tests

### type 

Type : **string/enum**

Can be `none`, `distant` or `amazon`

* `none` no deployment is made
* `local` runs a sub-process with lorhammer on the same consul that has the current orchestrator
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

You can launch orchestrator in cli mode to have some utilities (stop current scenarios, shutdown lorhammers, count lorhammers...)

```shell
ochestrator -consul 127.0.0.1:8500 -cli
```
