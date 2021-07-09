#!/bin/bash
# set -e
echo "Deploying support services..."
source ${ROOT_DIR}/scripts/env_vars.sh

source ${ROOT_DIR}/scripts/deploy_prom.sh

# RabbitMQ
# docker service rm rabbitmq
# docker build -t ${REPOSITORY}/rabbitmq ${SUPPORT_DIR}/rabbitmq/
# docker service create --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 ${REPOSITORY}/rabbitmq

# Spring Cloud Configuration Server
# cd support/config-server
# ./gradlew -Dhttp.proxyHost=openwrt.internal -Dhttp.proxyPort=1082 -Dhttps.proxyHost=openwrt.internal -Dhttps.proxyPort=1082 build
# cd ../..
# docker build -t ${REPOSITORY}/configserver support/config-server/
# docker service rm configserver
# docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 ${REPOSITORY}/configserver

# Hystrix Dashboard
#docker build -t ${REPOSITORY}/hystrix support/monitor-dashboard
#docker service rm hystrix
#docker service create --constraint node.role==manager --replicas 1 -p 7979:7979 --name hystrix --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 ${REPOSITORY}/hystrix

# Turbine
#docker service rm turbine
#docker service create --constraint node.role==manager --replicas 1 -p 8382:8282 --name turbine --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/turbine

# Zipkin
# docker service rm zipkin
# docker service create \
# 	--constraint node.role==manager --replicas 1 \
# 	-p 9411:9411 --name zipkin --network my_network \
# 	--update-delay 10s --with-registry-auth  \
# 	--update-parallelism 1 openzipkin/zipkin-slim

# Edge server
# docker service rm edge-server
#docker service create \
#	--replicas 1 --name edge-server -p 8765:8765 \
#	--network my_network --update-delay 10s --with-registry-auth \
#	 --update-parallelism 1 eriklupander/edge-server
#cd support/edge-server
#./gradlew -Dhttp.proxyHost=openwrt.internal -Dhttp.proxyPort=1082 -Dhttps.proxyHost=openwrt.internal -Dhttps.proxyPort=1082 clean build
#cd ../..
#docker build -t ${REPOSITORY}/edge-server support/edge-server/
#docker service rm edge-server
#docker service create --replicas 1 --name configserver -p 8765:8765 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 ${REPOSITORY}/edge-server
