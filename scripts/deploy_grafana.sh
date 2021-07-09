#!/bin/bash

# create a persistent volume for your data in /var/lib/grafana (database and plugins)
docker volume create grafana-storage

# Docke swarm mode grafana
docker service create -p 3000:3000 --constraint node.role==manager --name=grafana --replicas=1 --network=my_network \
    --mount type=volume,source=grafana-storage,target=/var/lib/grafana \
    grafana/grafana

# docker contaienr mode grafana
docker run -d -p 3000:3000 --name=grafana -v grafana-storage:/var/lib/grafana grafana/grafana