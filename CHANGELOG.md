# Changelog

## Version 0.7.0 - 2018-07-18

* **DEP** Update [issues/80](https://gitlab.com/itk.fr/lorhammer/issues/80) :
  * golang to 1.10.1
  * docker-compose to 1.20.1
  * hugo to 0.38.1 and repair site static
* **SCENARIO** update examples scenario [issues/75](https://gitlab.com/itk.fr/lorhammer/issues/75)
* **PULL_RESP** : proper support of `PULL_RESP` with `TX_ACK` [issues/81](https://gitlab.com/itk.fr/lorhammer/issues/81)
  * Add tests on *gateway* part
* **PROMETHEUS** : change of metrics name and add new to count pushack and pullresp [issues/80](https://gitlab.com/itk.fr/lorhammer/issues/80)
  * *lorhammer_pushack_durations* : Lora push ack latency distributions.
  * *lorhammer_pullresp_durations* : Lora pull resp latency distributions.
  * *lorhammer_pushack_long_request* : Lora nb lora push ack request witch take more than 2sc.
  * *lorhammer_pullresp_long_request* : Lora nb lora pull resp request witch take more than 2sc.
* **PROMETHEUS** : add metric for mqtt checker to be able to count messages after loraserver [issues/84](https://gitlab.com/itk.fr/lorhammer/issues/84)
  * *orchestrator_mqtt_ok* : Count of validated MQTT messages
  * *orchestrator_mqtt_failed* : Count of validation failed MQTT messages  
* **DEPLOYER** add multi ssh deployer [issues/85](https://gitlab.com/itk.fr/lorhammer/issues/85)
* **PROVISIONER** support new Brocaar provisioning API [issues/86](https://gitlab.com/itk.fr/lorhammer/issues/86)

## Version 0.6.0 - 2018-03-14

* **SCENARIO** new feature : launch multiple scenarios (payloads) with 1 orchestrator [issues/69](https://gitlab.com/itk.fr/lorhammer/issues/69) :
  * **BREAKING** before scenario file have 'init' property now its an array of init
  * **BREAKING** remove 'sleepAfterDeployTime' in deploy
  * **BREAKING** add 'requieredLorhammer' and 'maxWaitLorhammerTime' in scenario file
  * **BREAKING** remove ramp tytpe test because too complexe to manage in multiple scenarios
  * add a flag to lorhammer **--max-wait-orchestrator** to ask for an orchestrator for a max period (1 minute by default)
* **CONSUL** remove consul and pass mqtt url directly, permit to deploy more easily and to launch multiple orchestrator [issues/66](https://gitlab.com/itk.fr/lorhammer/issues/66)
* **LORHAMMER** allow to set port used to be scrapped by prometheus [issues/64](https://gitlab.com/itk.fr/lorhammer/issues/64)
* **ORCHESTRATOR** exit(1) if an error occur during scenario (previously exit > 0 only when checker fail) [issues/67](https://gitlab.com/itk.fr/lorhammer/issues/67) + repair bug on ramp type test
* **CHECKER** fix bug mqtt checker not working [issues/62](https://gitlab.com/itk.fr/lorhammer/issues/62)
* **DEP** Update :
  * dependencies
  * golang to 1.10
  * dep to 0.4.1
  * gitlab-ci to use artifact for vendor
  * rename packages **testType** and **testSuite** into **testype** and **testsuite** for lint
* **MAKE** `make help` display all tasks available with description
* **PROVISIONER** delete semtech v4 [issues/58](https://gitlab.com/itk.fr/lorhammer/issues/58)

## Version 0.5.0 - 2017-11-23 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.5.0)

* **PROVISIONER** increase timeout for http requests [issues/57](https://gitlab.com/itk.fr/lorhammer/issues/57)
* **CHECKER** add mqtt checker [issues/50](https://gitlab.com/itk.fr/lorhammer/issues/50)
* **MQTT** fix bug : mosquitto refuse to talk to a client with clientId containing a / [issues/59](https://gitlab.com/itk.fr/lorhammer/issues/59)
* **CHECKER** Remove consulClient in checker interface [issues/55](https://gitlab.com/itk.fr/lorhammer/issues/55)

## Version 0.4.0 - 2017-11-17 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.4.0)

* **CI** use golint [issues/49](https://gitlab.com/itk.fr/lorhammer/issues/49)
* **SITE** add badge from [https://goreportcard.com/report/gitlab.com/itk.fr/lorhammer](goreportcard.com)
* **SITE** fix font url
* **SCENARIO** Use 'date' property of payload as Rxpk Date [issues/53](https://gitlab.com/itk.fr/lorhammer/issues/53)
* **EXAMPLE** Add working example of lorhammer against [loraserver](https://www.loraserver.io/) [issues/52](https://gitlab.com/itk.fr/lorhammer/issues/52)
* **PROVISIONER** Create [loraserver](https://www.loraserver.io/) organization during lorhammer tests [issues/1](https://gitlab.com/itk.fr/lorhammer/issues/1)
* **CLEAN** Delete unwanted `.vscode` directory [issues/54](https://gitlab.com/itk.fr/lorhammer/issues/54)

## Version 0.3.1 - 2017-10-23 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.3.1)

* **PROVISIONER** Fix bug : when HTTP generic provisioner Deprovision() bad scenario was send [issues/51](https://gitlab.com/itk.fr/lorhammer/issues/51)

## Version 0.3.0 - 2017-10-19 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.3.0)

* **SCENARIO** Add a property to set the number of times the entire set of payloads will be sent [issues/48](https://gitlab.com/itk.fr/lorhammer/issues/48)
* **CI** use make to build [issues/46](https://gitlab.com/itk.fr/lorhammer/issues/46)
* **GO** update to golang 1.9 and use sync.map instead of concurrent-map [issues/47](https://gitlab.com/itk.fr/lorhammer/issues/47)
* **PROVISIONER** Add HTTP generic provisioner with a POST to add sensors and POST to delete them [issues/38](https://gitlab.com/itk.fr/lorhammer/issues/38)

## Version 0.2.2 - 2017-07-27 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.2.2)

* **CI** repair docker creation on tag

## Version 0.2.1 - 2017-07-27 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.2.1)

* **CI** clean artifact released by deleting useless i386 arch and directories containing binaries already present in tar.gz [issues/41](https://gitlab.com/itk.fr/lorhammer/issues/41)
* **ORCHESTRATOR** fix bug : kafka receive message after closing kafka connexion [issues/42](https://gitlab.com/itk.fr/lorhammer/issues/42)
* **LORHAMMER** fix bug : if no ip found lorhammer crash with index out of range [issues/43](https://gitlab.com/itk.fr/lorhammer/issues/43)
* **ORCHESTRATOR** fix bug : if a scenario fail, but next scenario success, orchestrator exit(>0) now [issues/25](https://gitlab.com/itk.fr/lorhammer/issues/25)
* **PROVISIONER** Goroutine over node creation to speed process in [loraserver](https://www.loraserver.io/) provisioner, you need to add `nbProvisionerParallel: 10` in scenario file [issues/17](https://gitlab.com/itk.fr/lorhammer/issues/17)

## Version 0.2.0 - 2017-07-21 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.2.0)

> Scenario file format has changed, please update it before use new lorhammer version!

* **CI** Use [goreleaser](https://github.com/goreleaser/goreleaser) to have better deployment management [issues/32](https://gitlab.com/itk.fr/lorhammer/issues/32)
* **SCENARIO** A payload can have a specific date [issues/37](https://gitlab.com/itk.fr/lorhammer/issues/37)
* **SCENARIO** Specify time before considering a no ack received in error [issues/31](https://gitlab.com/itk.fr/lorhammer/issues/31)
* **CI** Deploy lorhammer docker version [issue/26](https://gitlab.com/itk.fr/lorhammer/issues/26)
* **CODE** Use [golang/dep](https://github.com/golang/dep) to manage dependencies [issue/36](https://gitlab.com/itk.fr/lorhammer/issues/36)
* **DEPLOYER** Local deployer can set local-ip for deployed lorhammers [issue/23](https://gitlab.com/itk.fr/lorhammer/issues/23)
* **SCENARIO** Time before check [issue/24](https://gitlab.com/itk.fr/lorhammer/issues/24)
* **SCENARIO** Specify rxpk date to send [issue/22](https://gitlab.com/itk.fr/lorhammer/issues/22)
* **CHECKER** Kafka checker allow to check only sub-part of a message [issue/21](https://gitlab.com/itk.fr/lorhammer/issues/21)
* **SCENARIO** SleepTime between scenario is configurable [issue/2](https://gitlab.com/itk.fr/lorhammer/issues/2)
* **PROVISIONER** Use existing application to provision [loraserver](https://www.loraserver.io/) [issue/18](https://gitlab.com/itk.fr/lorhammer/issues/18)
* **CHECKER** Checker kafka : listen kafka queue and check message [issue/14](https://gitlab.com/itk.fr/lorhammer/issues/14)
* **CHECKER** Abstract checker concept [issue/15](https://gitlab.com/itk.fr/lorhammer/issues/15)

## Version 0.1.0 - 2017-05-03 - [binaries](https://gitlab.com/itk.fr/lorhammer/tags/0.1.0)

First version, lorhammer is intensively develop. **We do not care about api change.** Use lorhammer with caution.

Features :

* Can stress a lorawan network-server, through scenarios
* Can be distributed
* An orchestrator can manage lot of lorhammers
* Can deploy lorhammer over ssh and over amazon
* Can check if result are good over prometheus api call
* Can display what append over grafana