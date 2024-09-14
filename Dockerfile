FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 ldflags="-s -w" go build -o /iceperf-agent cmd/iceperf/main.go

FROM gcr.io/distroless/static

WORKDIR /

COPY --from=builder /iceperf-agent .

ENTRYPOINT ["./iceperf-agent"]
CMD ["-h"]
