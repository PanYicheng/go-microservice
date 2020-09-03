#!/bin/bash
# Create network for connecting local services
docker network create --driver overlay --attachable my_network
