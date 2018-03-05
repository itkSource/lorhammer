# Run lorhammer against loraserver

* Launch loraserver

> ADVERTIZED_HOST=YOUR_LOCAL_IP docker-compose up

* Launch lorhammer

> ./build/orchestrator -from-file ./resources/examples/loraserver/scenario.json -mqtt tcp://127.0.0.1:1883

* Open [prometheus](http://127.0.0.1:9090)
* You can enter request to have some feedback :
    * Number of lorawan request witch receive an ACK before 1 minute `sum(delta(lorhammer_durations_bucket{le=\"+Inf\"}[1m]))`