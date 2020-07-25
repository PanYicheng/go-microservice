#!/bin/bash
export GOOS=linux
export GOARCH=amd64
# Set to use static linking
export CGO_ENABLED=0

# Build accountservice
cd accountservice;go get;go build -o accountservice-$GOOS-$GOARCH; echo built `pwd`;cd ..
# builds the healthchecker binary
cd healthchecker;go get;go build -o healthchecker-$GOOS-$GOARCH; echo built `pwd`;cd ..
cp healthchecker/healthchecker-$GOOS-$GOARCH accountservice/
docker build -t unusedprefix/accountservice accountservice/

# Build vipservice
cd vipservice;go get;go build -o vipservice-$GOOS-$GOARCH; echo built `pwd`;cd ..
# builds the healthchecker binary
cp healthchecker/healthchecker-$GOOS-$GOARCH vipservice/

docker build -t unusedprefix/vipservice vipservice/

docker service rm quotes-service
docker service rm accountservice
docker service rm vipservice

GELF_ADDRESS=udp://192.168.50.3:12202
docker service create \
	--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
	--log-opt gelf-compression-type=none \
   	--name=quotes-service --replicas=1 --network=my_network eriklupander/quotes-service
docker service create \
	--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
	--log-opt gelf-compression-type=none \
	--name=accountservice --replicas=1 --network=my_network -p=6767:6767 unusedprefix/accountservice
docker service create \
	--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
	--log-opt gelf-compression-type=none \
	--name=vipservice --replicas=1 --network=my_network unusedprefix/vipservice

#docker service create \
#   	--name=quotes-service --replicas=1 --network=my_network eriklupander/quotes-service
#docker service create \
#	--name=accountservice --replicas=1 --network=my_network -p=6767:6767 unusedprefix/accountservice
#docker service create \
#	--name=vipservice --replicas=1 --network=my_network unusedprefix/vipservice
