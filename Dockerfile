FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /iceperf-agent cmd/iceperf/main.go

FROM alpine

WORKDIR /

COPY --from=builder /iceperf-agent .

RUN ls -lsa /

ENTRYPOINT ["./iceperf-agent"]
CMD ["-config", "config.yaml"]
