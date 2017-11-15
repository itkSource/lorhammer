# Run lorhammer against loraserver

* Launch tools

> LORHAMMER_PROMETHEUS_IP="YOUR_IP" LORHAMMER_MQTT_IP="YOUR_IP" LORHAMMER_MQTT_PORT="1884" LORHAMMER_CONSUL_IP="YOUR_IP" LORHAMMER_GRAFANA_IP="YOUR_IP" ./resources/scripts/launchTools.sh

* Launch loraserver

> docker-compose up

* Launc lorhammer

> ./build/orchestrator -from-file ./resources/examples/loraserver/scenario.json -consul 127.0.0.1:8500

* Open [grafana](http://127.0.0.1:3000)

You must see something like :

![grafana screenshot](screenshotGrafana)