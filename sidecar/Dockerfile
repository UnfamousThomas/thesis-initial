FROM golang:1.23.0 as builder

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOARCH=arm64 go build -a -o sidecar cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/sidecar .
USER 65532:65532

ENTRYPOINT ["/sidecar"]