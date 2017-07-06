#!/bin/bash

set -o errexit
set -o pipefail

if [ -z "$LORHAMMER_MQTT_IP" ]; then
    echo "Need to set LORHAMMER_MQTT_IP"
    exit 1
fi

if [ -z "$LORHAMMER_PROMETHEUS_IP" ]; then
    echo "Need to set LORHAMMER_PROMETHEUS_IP"
    exit 1
fi

if [ -z "$LORHAMMER_CONSUL_IP" ]; then
    echo "Need to set LORHAMMER_CONSUL_IP"
    exit 1
fi

if [ -z "$LORHAMMER_GRAFANA_IP" ]; then
    echo "Need to set LORHAMMER_GRAFANA_IP"
    exit 1
fi

ARGS=()
if [ ! -z "$CI_BUILD_ID" ]; then
    ARGS+=( '-f' 'docker-compose.yml' '-f' 'docker-compose.integration.yml' '-p' "${CI_BUILD_ID}" )
fi

echo "Start launchTools.sh with LORHAMMER_MQTT_IP=${LORHAMMER_MQTT_IP} LORHAMMER_PROMETHEUS_IP=${LORHAMMER_PROMETHEUS_IP} LORHAMMER_CONSUL_IP=${LORHAMMER_CONSUL_IP} LORHAMMER_GRAFANA_IP=${LORHAMMER_GRAFANA_IP}"
echo "Clean docker deamon : docker-compose ${ARGS[@]} down --remove-orphans"
docker-compose ${ARGS[@]} down --remove-orphans
echo "Start consul : docker-compose ${ARGS[@]} up -d consul"
docker-compose ${ARGS[@]} up -d consul
echo "force to build images"
docker-compose ${ARGS[@]} build
echo "Sleep 20sc to lets consul start"
sleep 20
echo "Start tools and register them into consul : docker-compose ${ARGS[@]} up -d mqtt prometheus grafana cadvisor consul-register"
docker-compose ${ARGS[@]} up -d mqtt prometheus grafana cadvisor consul-register
