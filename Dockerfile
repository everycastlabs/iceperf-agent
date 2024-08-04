FROM golang:1.22.5-bookworm AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /iceperf-agent cmd/iceperf/main.go

FROM debian:bookworm-slim

WORKDIR /

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /iceperf-agent .

RUN ls -lsa /
COPY config-api.yaml config.yaml

ENTRYPOINT ["./iceperf-agent"]
CMD ["-config", "config.yaml"]
