
SERVER_MAIN := cmd/server/server.go

SERVER_BINARY := bin/chinchilla-server
RELEASE_ZIP := release/chinchilla-server.zip

.PHONY: deps release test mockgen

release: $(RELEASE_ZIP)

test:
	go test ./...

$(SERVER_BINARY): deps $(SERVER_MAIN) proto/agent.pb.go
	go build -ldflags="-X main.version=$(SERVER_VERSION)" -o $@ $(SERVER_MAIN)

proto/agent.pb.go: proto/agent.proto
	protoc --go_out=plugins=grpc:. $<

$(RELEASE_ZIP): $(SERVER_BINARY)
	mkdir -p release
	zip $@ $<

deps:
	glide install
	rm -rf vendor/github.com/docker/docker/vendor

mockgen:
	mockgen -destination mocks/mock_server.go -package mocks github.com/Trojan295/chinchilla-server/server AgentStore,GameserverStore
