#!/bin/bash

cd support/config-server
./gradlew -Dhttp.proxyHost=192.168.50.4 -Dhttp.proxyPort=1082 -Dhttps.proxyHost=192.168.50.4 -Dhttps.proxyPort=1082 build
cd ../..
docker build -t unusedprefix/configserver support/config-server/
docker service rm configserver
docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 unusedprefix/configserver

# Hystrix Dashboard
docker build -t unusedprefix/hystrix support/monitor-dashboard
docker service rm hystrix
docker service create --constraint node.role==manager --replicas 1 -p 7979:7979 --name hystrix --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 unusedprefix/hystrix
# Turbine
docker service rm turbine
docker service create --constraint node.role==manager --replicas 1 -p 8382:8282 --name turbine --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/turbine
