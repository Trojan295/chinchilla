FROM golang:1.13-alpine AS builder

COPY . /go/src/github.com/Trojan295/chinchilla
WORKDIR /go/src/github.com/Trojan295/chinchilla

RUN apk update \
    && apk add glide make git \
    && make bin/chinchilla-agent

FROM alpine

RUN apk update && apk add gettext
COPY --from=builder /go/src/github.com/Trojan295/chinchilla/bin/chinchilla-agent /chinchilla-agent
COPY docker/chinchilla.toml.tmpl docker/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/chinchilla-agent"]
