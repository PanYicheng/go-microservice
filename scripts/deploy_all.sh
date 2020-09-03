#!/bin/bash
echo "Deploying all microservices..."
# ROOT_DIR="$(dirname `pwd`)"
ROOT_DIR=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice
source ${ROOT_DIR}/scripts/env_vars.sh

# Deploy accountservice
${ROOT_DIR}/scripts/deploy_accountservice.sh

# Go Build imageservice
${ROOT_DIR}/build/build_imageservice.sh
# Go Build vipservice
${ROOT_DIR}/build/build_vipservice.sh

# Remove services before creating new ones
echo "Removing old services: quotes-service, vipservice, imageservice"
docker service rm quotes-service
docker service rm vipservice
docker service rm imageservice

#GELF_ADDRESS=udp://192.168.50.3:12202
ENVIRONMENT=test

docker service create \
	--name=quotes-service --replicas=1 --network=my_network \
	eriklupander/quotes-service

if [ ! -z ${GELF_ADDRESS+x}] 
then
# if gelftail enabled
	docker service create \
		--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
		--log-opt gelf-compression-type=none \
		--name=vipservice --replicas=1 \
		--network=my_network \
		--env ENVIRONMENT=${ENVIRONMENT} \
		--env SERVICE_NAME=vipservice \
		--env SERVICE_PORT=6767 \
		--env ZIPKIN_URL=http://zipkin:9411 \
		--env AMQP_URL=amqp://guest:guest@rabbitmq:5672 \
		unusedprefix/vipservice
	docker service create \
		--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
		--log-opt gelf-compression-type=none \
		--name=imageservice --replicas=1 \
		--network=my_network \
		--env ENVIRONMENT=${ENVIRONMENT} \
		--env SERVICE_NAME=imageservice \
		--env SERVICE_PORT=7777 \
		--env ZIPKIN_URL=http://zipkin:9411 \
		unusedprefix/imageservice
else
# if gelftail not enabled
	docker service create \
		--name=vipservice --replicas=1 \
		--network=my_network \
		--env ENVIRONMENT=${ENVIRONMENT} \
		--env SERVICE_NAME=vipservice \
		--env SERVICE_PORT=6767 \
		--env ZIPKIN_URL=http://zipkin:9411 \
		--env AMQP_URL=amqp://guest:guest@rabbitmq:5672 \
		unusedprefix/vipservice
	docker service create \
		--name=imageservice --replicas=1 \
		--network=my_network \
		--env ENVIRONMENT=${ENVIRONMENT} \
		--env SERVICE_NAME=imageservice \
		--env SERVICE_PORT=7777 \
		--env ZIPKIN_URL=http://zipkin:9411 \
		unusedprefix/imageservice
fi
