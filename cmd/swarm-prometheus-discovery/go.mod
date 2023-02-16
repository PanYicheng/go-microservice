module github.com/PanYicheng/go-microservice/cmd/swarm-prometheus-discovery

go 1.14

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

require (
	github.com/alexflint/go-arg v1.3.1-0.20200806235247-96c756c382ed
	github.com/containerd/containerd v1.5.18 // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
)
