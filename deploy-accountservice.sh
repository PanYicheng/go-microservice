#!/bin/bash
echo Build gobinaries and container images...
export GOOS=linux
export GOARCH=amd64
# Set to use static linking
export CGO_ENABLED=0

# Go Build accountservice
cd accountservice;go get;go build -o accountservice-$GOOS-$GOARCH; echo "  built `pwd`";cd ..
# Go Build healthchecker binary
# cd healthchecker;go get;go build -o healthchecker-$GOOS-$GOARCH; echo "  built `pwd`";cd ..

# Docker Build accountservice
cp healthchecker/healthchecker-$GOOS-$GOARCH accountservice/
docker build -t unusedprefix/accountservice accountservice/

# Remove services before creating new ones
echo Removing old services...
docker service rm accountservice

#GELF_ADDRESS=udp://192.168.50.3:12202
if [ ! -z ${GELF_ADDRESS+x}] 
then
# if gelftail enabled
	docker service create \
		--log-driver=gelf --log-opt gelf-address=$GELF_ADDRESS \
		--log-opt gelf-compression-type=none \
		--name=accountservice --replicas=1 --network=my_network -p=6767:6767 unusedprefix/accountservice
else
# if gelftail not enabled
	docker service create \
		--name=accountservice --replicas=1 --network=my_network -p=6767:6767 unusedprefix/accountservice
fi
