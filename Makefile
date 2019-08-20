
SERVER_MAIN := cmd/server.go

.PHONY: deps release

release: chinchilla-server.zip

chinchilla-server: deps $(SERVER_MAIN)
	go build -ldflags="-X main.version=$(SERVER_VERSION)" -o $@ $(SERVER_MAIN)

chinchilla-server.zip: chinchilla-server
	zip $@ $<

deps:
	glide install
	rm -rf vendor/github.com/docker/docker/vendor