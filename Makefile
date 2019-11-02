SERVER_MAIN := cmd/server/server.go
SERVER_BINARY := bin/chinchilla-server

SCHEDULER_MAIN := cmd/scheduler/scheduler.go
SCHEDULER_BINARY := bin/chinchilla-scheduler

BINARIES := $(SERVER_BINARY) $(SCHEDULER_BINARY)
RELEASE_FILES := chinchilla.toml chinchilla-server.service

RELEASE_ZIP := release/chinchilla.zip

.PHONY: deps release test mockgen

release: $(RELEASE_ZIP)

$(RELEASE_ZIP): $(BINARIES)
	mkdir -p release
	zip $@ $(BINARIES) $(RELEASE_FILES)

$(SERVER_BINARY): deps $(SERVER_MAIN) proto/agent.pb.go
	go build -ldflags="-X main.version=$(VERSION)" -o $@ $(SERVER_MAIN)

$(SCHEDULER_BINARY): deps $(SCHEDULER_MAIN) proto/agent.pb.go
	go build -ldflags="-X main.version=$(VERSION)" -o $@ $(SCHEDULER_MAIN)

proto/agent.pb.go: proto/agent.proto
	protoc --go_out=plugins=grpc:. $<

test:
	go test ./...

deps:
	glide install
	rm -rf vendor/github.com/docker/docker/vendor

mockgen:
	mockgen -destination mocks/mock_server.go -package mocks github.com/Trojan295/chinchilla/server AgentStore,GameserverStore
