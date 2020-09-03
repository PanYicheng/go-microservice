#!/bin/bash

# Remove main services
docker service rm accountservice
docker service rm quotes-service
docker service rm vipservice
docker service rm imageservice
# Remove springcloud & support services
docker service rm configserver
docker service rm rabbitmq
docker service rm gelftail
docker service rm viz
docker service rm hystrix
docker service rm turbine
docker service rm zipkin
docker service rm edge-server
docker service rm grafana
docker service rm prometheus
docker service rm swarm-prometheus-discovery

# docker network rm my_network
