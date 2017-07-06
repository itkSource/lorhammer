#!/bin/bash

go get -u -v github.com/Sirupsen/logrus
go get -u -v github.com/prometheus/client_golang/prometheus
go get -u -v github.com/jacobsa/crypto
go get -u -v github.com/hashicorp/consul/api
go get -u -v github.com/eclipse/paho.mqtt.golang
#TODO FIXME https://github.com/eclipse/paho.mqtt.golang/issues/121
cd $GOPATH/src/github.com/eclipse/paho.mqtt.golang && git checkout c37a0a2 && cd -
go get -u -v github.com/google/uuid
go get -u -v github.com/orcaman/concurrent-map
go get -u -v golang.org/x/net/websocket
go get -u -v golang.org/x/sys/windows
go get -u -v github.com/brocaar/lorawan
go get -u -v github.com/aws/aws-sdk-go
go get -u -v github.com/brocaar/lora-gateway-bridge/gateway/
go get -u -v github.com/Shopify/sarama