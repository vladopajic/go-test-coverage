# ===============================================================
# GO_VERSION: automatically update to most recent via dependabot
FROM golang:1.27.1 AS builder
WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download all

COPY ./ ./

# TOOL_VERSION: when changing version update version in other places
ARG VERSION=v2.22.0
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-X main.Version=${VERSION}" -o go-test-coverage .

RUN chmod +x docker-entrypoint.sh

# ===============================================================
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /

COPY --from=builder /workspace/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=builder /workspace/go-test-coverage /go-test-coverage
ENTRYPOINT ["/docker-entrypoint.sh"]
