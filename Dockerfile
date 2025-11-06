# GO_VERSION: automatically update to most recent via dependabot
FROM --platform=$BUILDPLATFORM golang:1.25.3 AS builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download all

COPY ./ ./

ARG VERSION
ARG TARGETOS 
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o go-test-coverage .

FROM --platform=$BUILDPLATFORM gcr.io/distroless/base:latest
WORKDIR /
COPY --from=builder /workspace/go-test-coverage .
COPY --from=builder /usr/local/go/bin/go /usr/local/go/bin/go
ENV PATH="${PATH}:/usr/local/go/bin"
ENTRYPOINT ["/go-test-coverage"]
