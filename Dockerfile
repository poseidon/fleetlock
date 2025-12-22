FROM docker.io/golang:1.25.5 AS builder
COPY . src
RUN cd src && make build

FROM docker.io/alpine:3.23.2
LABEL maintainer="Dalton Hubble <dghubble@gmail.com>"
LABEL org.opencontainers.image.title="fleetlock",
LABEL org.opencontainers.image.source="https://github.com/poseidon/fleetlock"
LABEL org.opencontainers.image.vendor="Poseidon Labs"
RUN apk --no-cache --update add ca-certificates
COPY --from=builder /go/src/bin/fleetlock /usr/local/bin
USER nobody
ENTRYPOINT ["/usr/local/bin/fleetlock"]
