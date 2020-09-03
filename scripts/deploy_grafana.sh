#!/bin/bash
docker service create -p 3000:3000 --constraint node.role==manager --name=grafana --replicas=1 --network=my_network grafana/grafana
