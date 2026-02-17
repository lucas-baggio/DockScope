module github.com/dockscope/dockscope

go 1.24.0

require (
	github.com/docker/docker v20.10.27+incompatible
	github.com/gorilla/websocket v1.5.3
)

replace github.com/docker/distribution => github.com/distribution/distribution v2.8.2+incompatible

require (
	github.com/Microsoft/go-winio v0.4.21 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)
