module github.com/PanYicheng/go-microservice/cmd/swarm-prometheus-discovery

go 1.14

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/alexflint/go-arg v1.3.1-0.20200806235247-96c756c382ed
	github.com/containerd/containerd v1.4.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	google.golang.org/grpc v1.32.0 // indirect
)
