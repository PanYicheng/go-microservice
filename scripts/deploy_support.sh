#!/bin/bash
# set -e

SUPPORT_DIR=../deployments
# RabbitMQ
docker service rm rabbitmq
docker build -t unusedprefix/rabbitmq ${SUPPORT_DIR}/rabbitmq/
docker service create --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 unusedprefix/rabbitmq

# Spring Cloud Configuration Server
# cd support/config-server
# ./gradlew -Dhttp.proxyHost=openwrt.internal -Dhttp.proxyPort=1082 -Dhttps.proxyHost=openwrt.internal -Dhttps.proxyPort=1082 build
# cd ../..
# docker build -t unusedprefix/configserver support/config-server/
# docker service rm configserver
# docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 unusedprefix/configserver

# Hystrix Dashboard
#docker build -t unusedprefix/hystrix support/monitor-dashboard
#docker service rm hystrix
#docker service create --constraint node.role==manager --replicas 1 -p 7979:7979 --name hystrix --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 unusedprefix/hystrix

# Turbine
#docker service rm turbine
#docker service create --constraint node.role==manager --replicas 1 -p 8382:8282 --name turbine --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/turbine

# Zipkin
docker service rm zipkin
docker service create \
	--constraint node.role==manager --replicas 1 \
	-p 9411:9411 --name zipkin --network my_network \
	--update-delay 10s --with-registry-auth  \
	--update-parallelism 1 openzipkin/zipkin-slim

# Edge server
# docker service rm edge-server
#docker service create \
#	--replicas 1 --name edge-server -p 8765:8765 \
#	--network my_network --update-delay 10s --with-registry-auth \
#	 --update-parallelism 1 eriklupander/edge-server
#cd support/edge-server
#./gradlew -Dhttp.proxyHost=openwrt.internal -Dhttp.proxyPort=1082 -Dhttps.proxyHost=openwrt.internal -Dhttps.proxyPort=1082 clean build
#cd ../..
#docker build -t unusedprefix/edge-server support/edge-server/
#docker service rm edge-server
#docker service create --replicas 1 --name configserver -p 8765:8765 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 unusedprefix/edge-server

# Prometheus
docker build -t unusedprefix/prometheus ${SUPPORT_DIR}/prometheus
docker service rm prometheus
docker service create -p 9090:9090 --constraint node.role==manager --mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local --name=prometheus --replicas=1 --network=my_network unusedprefix/prometheus

# swarm-prometheus-discovery
# Set to use static linking
export CGO_ENABLED=0
cd ../cmd/swarm-prometheus-discovery;go get;go build;echo built `pwd`;cd ../../scripts
cp ../cmd/swarm-prometheus-discovery/swarm-prometheus-discovery ${SUPPORT_DIR}/swarm-prometheus-discovery/

docker build -t unusedprefix/swarm-prometheus-discovery ${SUPPORT_DIR}/swarm-prometheus-discovery/
docker service rm swarm-prometheus-discovery
docker service create \
	--constraint node.role==manager \
	--mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local \
	--mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
	--name=swarm-prometheus-discovery --replicas=1 --network=my_network \
	--env NETWORK_NAME=my_network \
	--env IGNORED_SERVICE_STR=prometheus,grafana \
	unusedprefix/swarm-prometheus-discovery
