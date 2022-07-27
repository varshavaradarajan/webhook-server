FROM golang:1.16 as builder
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/

# Copy the go source
COPY cmd/server/main.go main.go
RUN GOFLAGS="-mod=vendor" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
 go build -a -o server main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/server .
USER nonroot:nonroot

ENTRYPOINT ["/server"]
