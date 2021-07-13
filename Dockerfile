FROM docker.io/golang:1.17beta1 AS builder
COPY . src
RUN cd src && make build

FROM docker.io/alpine:3.14.0
LABEL maintainer="Dalton Hubble <dghubble@gmail.com>"
RUN apk --no-cache --update add ca-certificates
COPY --from=builder /go/src/bin/fleetlock /usr/local/bin
USER nobody
ENTRYPOINT ["/usr/local/bin/fleetlock"]
