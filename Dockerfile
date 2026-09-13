# syntax=docker/dockerfile:1
# check=error=true

# Latest version: https://hub.docker.com/_/golang/tags
FROM --platform=$BUILDPLATFORM golang:1.27.1-trixie AS base

WORKDIR /src

RUN apt-get update \
    && apt-get install --assume-yes --no-install-recommends \
        ca-certificates \
        tree \
        git \
        openssh-client

FROM base AS builder-download

COPY go.mod .
COPY go.sum .

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM builder-download AS build

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG GOOS
ARG GOARCH
ARG GO_MODULE=github.com/specsnl/specsdeployd
ARG SPECSDEPLOYD_VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${GOOS:-$TARGETOS} GOARCH=${GOARCH:-$TARGETARCH} go build \
        -trimpath \
        -tags netgo \
        -ldflags "-s -w -X ${GO_MODULE}/internal/cmd.Version=${SPECSDEPLOYD_VERSION}" -o ./specsdeployd

# No runtime image: specsdeployd is a host systemd service driving podman and
# systemctl, so this Dockerfile only ever produces the binary.
FROM scratch AS export

COPY --from=build /src/specsdeployd /specsdeployd
