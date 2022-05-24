ARG base="golang:1.18.2"

# Build the manager binary
FROM ${base} AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY ./go.mod ./go.sum ./

# disable proxy
ENV GOPROXY=https://proxy.golang.org
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY *.go .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o app main.go metric_types.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/app .
USER 65532:65532

ENV LOG_LEVEL="info"
ENV TCP_ADDR="localhost:8123"
ENV METRICS_ADDR=":8080"
ENV METRICS_PATH="/metrics"

ENTRYPOINT ["/app"]
