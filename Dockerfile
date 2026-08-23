FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /bin/aegisbox ./cmd/aegisbox

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/aegisbox /usr/local/bin/aegisbox
COPY configs/config.yaml /etc/aegisbox/config.yaml
COPY scripts/setup-rootfs.sh /opt/aegisbox/scripts/setup-rootfs.sh

RUN chmod +x /opt/aegisbox/scripts/setup-rootfs.sh && \
    /opt/aegisbox/scripts/setup-rootfs.sh /opt/aegisbox/rootfs/python

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/aegisbox"]
CMD ["server", "-port", "8080", "-host", "0.0.0.0"]
