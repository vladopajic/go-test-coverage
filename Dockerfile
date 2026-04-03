# ===============================================================
# GO_VERSION: automatically update to most recent via dependabot
FROM golang:1.26.1 AS builder
WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download all

COPY ./ ./

ARG VERSION
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o go-test-coverage .
    
COPY docker-entrypoint.sh docker-entrypoint.sh
RUN chmod +x docker-entrypoint.sh

# ===============================================================
FROM debian:bookworm-slim
WORKDIR /

COPY --from=builder /workspace/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=builder /workspace/go-test-coverage .
COPY --from=builder /usr/local/go/bin/go /usr/local/go/bin/go
ENV PATH="${PATH}:/usr/local/go/bin"
ENTRYPOINT ["/docker-entrypoint.sh"]