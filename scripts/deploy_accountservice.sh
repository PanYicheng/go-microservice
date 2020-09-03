#!/bin/bash
echo "Deploying accountservice..."
# ROOT_DIR="$(dirname `pwd`)"
ROOT_DIR=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice
source ${ROOT_DIR}/scripts/env_vars.sh

# Build accountservice first.
${ROOT_DIR}/build/build_accountservice.sh

# Remove services before creating new ones
echo Removing old accountservice
docker service rm accountservice

if [ ! -z ${GELF_ADDRESS+x}] 
then
# if gelftail enabled
	echo "Using gelftail log server"
	docker service create \
		--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
		--log-opt gelf-compression-type=none \
		--name=accountservice --replicas=1 \
		--network=my_network -p=6767:6767 \
		--env ENVIRONMENT=${ENVIRONMENT} \
		--env SERVICE_NAME=accountservice \
		--env SERVICE_PORT=6767 \
		--env ZIPKIN_URL=http://zipkin:9411 \
		--env AMQP_URL=amqp://guest:guest@rabbitmq:5672 \
		unusedprefix/accountservice
else
# if gelftail not enabled
	docker service create \
		--name=accountservice --replicas=1 \
		--network=my_network -p=6767:6767 \
		--env ENVIRONMENT=${ENVIRONMENT} \
		--env SERVICE_NAME=accountservice \
		--env SERVICE_NAME=accountservice \
		--env SERVICE_PORT=6767 \
		--env ZIPKIN_URL=http://zipkin:9411 \
		--env AMQP_URL=amqp://guest:guest@rabbitmq:5672 \
		unusedprefix/accountservice
fi
