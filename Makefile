
SERVER_MAIN := cmd/server.go

SERVER_BINARY := bin/chinchilla-server
RELEASE_ZIP := release/chinchilla-server.zip

.PHONY: deps release

release: $(RELEASE_ZIP)

$(SERVER_BINARY): deps $(SERVER_MAIN)
	go build -ldflags="-X main.version=$(SERVER_VERSION)" -o $@ $(SERVER_MAIN)

$(RELEASE_ZIP): $(SERVER_BINARY)
	mkdir -p release
	zip $@ $<

deps:
	glide install
	rm -rf vendor/github.com/docker/docker/vendor