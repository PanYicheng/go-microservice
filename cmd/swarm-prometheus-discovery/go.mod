module github.com/PanYicheng/go-microservice/cmd/swarm-prometheus-discovery

go 1.14

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

require (
	github.com/alexflint/go-arg v1.3.1-0.20200806235247-96c756c382ed
	github.com/docker/docker v1.4.2-0.20191101170500-ac7306503d23
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/sirupsen/logrus v1.6.0
)
