# GO_VERSION: go version should match version in go.mod file
FROM golang:1.22.7 as builder
WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download all

COPY ./ ./

ARG VERSION
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o go-test-coverage .

FROM gcr.io/distroless/base:latest
WORKDIR /
COPY --from=builder /workspace/go-test-coverage .
COPY --from=builder /usr/local/go/bin/go /usr/local/go/bin/go
ENV PATH="${PATH}:/usr/local/go/bin"
ENTRYPOINT ["/go-test-coverage"]