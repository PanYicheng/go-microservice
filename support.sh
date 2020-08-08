#!/bin/bash
# set -e

# RabbitMQ
docker service rm rabbitmq
docker build -t unusedprefix/rabbitmq support/rabbitmq/
docker service create --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 unusedprefix/rabbitmq

# Spring Cloud Configuration Server
cd support/config-server
./gradlew -Dhttp.proxyHost=openwrt.internal -Dhttp.proxyPort=1082 -Dhttps.proxyHost=openwrt.internal -Dhttps.proxyPort=1082 build
cd ../..
docker build -t unusedprefix/configserver support/config-server/
docker service rm configserver
docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 unusedprefix/configserver

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
