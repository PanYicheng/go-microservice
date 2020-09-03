#!/bin/bash
# swarm-prometheus-discovery
export GOOS=linux
export GOARCH=amd64
# Set to use static linking
export CGO_ENABLED=0

cd swarm-prometheus-discovery;go get;go build -o swarm-prometheus-discovery-${GOOS}-${GOARCH};echo built `pwd`;cd ..

docker build -t unusedprefix/swarm-prometheus-discovery swarm-prometheus-discovery/
docker service rm swarm-prometheus-discovery
docker service create  --constraint node.role==manager --mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock --name=swarm-prometheus-discovery --replicas=1 --network=my_network unusedprefix/swarm-prometheus-discovery
