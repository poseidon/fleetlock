FROM docker.io/alpine:3.12
MAINTAINER Dalton Hubble <dghubble@gmail.com>

RUN apk --no-cache --update add ca-certificates
COPY bin/fleetlock /usr/local/bin

USER nobody
ENTRYPOINT ["/usr/local/bin/fleetlock"]

