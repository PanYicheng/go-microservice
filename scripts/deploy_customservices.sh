docker service rm servicec
docker service create --name=servicec --replicas=1 --network=my_network --mount type=bind,source=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice/deployments/customtopology/services/servicec,target=/data/ unusedprefix/customservice

docker service rm servicee
docker service create --name=servicee --replicas=1 --network=my_network --mount type=bind,source=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice/deployments/customtopology/services/servicee,target=/data/ unusedprefix/customservice

docker service rm serviceb
docker service create --name=serviceb --replicas=1 --network=my_network --mount type=bind,source=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice/deployments/customtopology/services/serviceb,target=/data/ unusedprefix/customservice

docker service rm serviced
docker service create --name=serviced --replicas=1 --network=my_network --mount type=bind,source=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice/deployments/customtopology/services/serviced,target=/data/ unusedprefix/customservice

docker service rm servicea
docker service create --name=servicea --replicas=1 --network=my_network -p=6767:6767 --mount type=bind,source=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice/deployments/customtopology/services/servicea,target=/data/ unusedprefix/customservice

