#!/bin/bash

if [ -z "$LORHAMMER_MQTT_IP" ]; then
    echo "Need to set LORHAMMER_MQTT_IP"
    exit 1
fi

if [ -z "$LORHAMMER_MQTT_PORT" ]; then
    echo "LORHAMMER_MQTT_PORT not set use default 1883"
    LORHAMMER_MQTT_PORT=1883
fi

if [ -z "$LORHAMMER_PROMETHEUS_IP" ]; then
    echo "Need to set LORHAMMER_PROMETHEUS_IP"
    exit 1
fi

if [ -z "$LORHAMMER_PROMETHEUS_PORT" ]; then
    echo "LORHAMMER_PROMETHEUS_PORT not set use default 9090"
    LORHAMMER_PROMETHEUS_PORT=9090
fi

if [ -z "$LORHAMMER_CONSUL_IP" ]; then
    echo "Need to set LORHAMMER_CONSUL_IP"
    exit 1
fi

if [ -z "$LORHAMMER_CONSUL_PORT" ]; then
    echo "LORHAMMER_CONSUL_PORT not set use default 8500"
    LORHAMMER_CONSUL_PORT=8500
fi

if [ -z "$LORHAMMER_GRAFANA_IP" ]; then
    echo "Need to set LORHAMMER_GRAFANA_IP"
    exit 1
fi

if [ -z "$LORHAMMER_GRAFANA_PORT" ]; then
    echo "LORHAMMER_GRAFANA_PORT not set use default 3000"
    LORHAMMER_GRAFANA_PORT=3000
fi

CONSUL_URL="$LORHAMMER_CONSUL_IP:$LORHAMMER_CONSUL_PORT/v1/agent/service/register"

###
# $1 id of service to register (must be unique)
# $2 name of service to register
# $3 port of service to register
# $4 address of service to register
###
function register {
    echo "register {\"ID\":\"$1\",\"Name\":\"$2\",\"Port\":$3,\"Address\":\"$4\"" ${CONSUL_URL}
    curl -X PUT -H "Content-Type: application/json" -d "{\"ID\":\"$1\",\"Name\":\"$2\",\"Port\":$3,\"Address\":\"$4\"}" ${CONSUL_URL}
}

register "mqtt" "mqtt" ${LORHAMMER_MQTT_PORT} ${LORHAMMER_MQTT_IP}
register "prometheus" "prometheus" ${LORHAMMER_PROMETHEUS_PORT} ${LORHAMMER_PROMETHEUS_IP}
register "grafana" "grafana" ${LORHAMMER_GRAFANA_PORT} ${LORHAMMER_GRAFANA_IP}
register "cadvisor" "cadvisor" 8085 "cadvisor"
