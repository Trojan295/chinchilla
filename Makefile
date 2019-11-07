SERVER_MAIN := cmd/server/server.go
SERVER_BINARY := bin/chinchilla-server
SERVER_IMAGE := trojan295/chinchilla-server
SERVER_DOCKERFILE := docker/server/Dockerfile

SCHEDULER_MAIN := cmd/scheduler/scheduler.go
SCHEDULER_BINARY := bin/chinchilla-scheduler
SCHEDULER_IMAGE := trojan295/chinchilla-scheduler
SCHEDULER_DOCKERFILE := docker/server/Dockerfile

AGENT_MAIN := cmd/agent/agent.go
AGENT_BINARY := bin/chinchilla-agent
AGENT_IMAGE := trojan295/chinchilla-agent
AGENT_DOCKERFILE := docker/server/Dockerfile

BINARIES := \
	$(SERVER_BINARY) \
	$(SCHEDULER_BINARY) \
	$(AGENT_BINARY)

IMAGES := \
	$(SERVER_IMAGE) \
	$(SCHEDULER_IMAGE) \
	$(AGENT_IMAGE)

RELEASE_FILES := \
	$(BINARIES) \
	chinchilla.toml \
	chinchilla-server.service \
	chinchilla-agent.service

RELEASE_ZIP := release/chinchilla.zip

.PHONY: deps release test mockgen $(IMAGES) docker docker_push

# ----- RELEASES -----

release: $(RELEASE_ZIP)

$(RELEASE_ZIP): $(RELEASE_FILES)
	mkdir -p release
	zip $@ $(RELEASE_FILES)

# ----- BINARIES -----

$(SERVER_BINARY): deps $(SERVER_MAIN) proto/agent.pb.go
	go build -ldflags="-X main.version=$(VERSION)" -o $@ $(SERVER_MAIN)

$(SCHEDULER_BINARY): deps $(SCHEDULER_MAIN) proto/agent.pb.go
	go build -ldflags="-X main.version=$(VERSION)" -o $@ $(SCHEDULER_MAIN)

$(AGENT_BINARY): deps $(AGENT_MAIN) proto/agent.pb.go
	go build -ldflags="-X main.version=$(VERSION)" -o $@ $(AGENT_MAIN)

proto/agent.pb.go: proto/agent.proto
	protoc --go_out=plugins=grpc:. $<

deps:
	glide install
	rm -rf vendor/github.com/docker/docker/vendor

# ----- DOCKER -----

$(SERVER_IMAGE):
	docker build -f $(SERVER_DOCKERFILE) -t $(SERVER_IMAGE):$(VERSION) .

$(SCHEDULER_IMAGE):
	docker build -f $(SCHEDULER_DOCKERFILE) -t $(SCHEDULER_IMAGE):$(VERSION) .

$(AGENT_IMAGE):
	docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE):$(VERSION) .

docker: $(IMAGES)

docker_push: docker
	docker push $(SERVER_IMAGE):$(VERSION)
	docker push $(SCHEDULER_IMAGE):$(VERSION)
	docker push $(AGENT_IMAGE):$(VERSION)

# ----- UTILS -----

test:
	go test ./...

mockgen:
	mockgen -destination mocks/mock_server.go -package mocks github.com/Trojan295/chinchilla/server AgentStore,GameserverStore
	mockgen -destination mocks/mock_gameservers.go -package mocks github.com/Trojan295/chinchilla/server/gameservers LogStore
	mockgen -destination mocks/mock_etcd_client.go -package mocks go.etcd.io/etcd/client KeysAPI
