# Use a specific alpine version for better reproducibility
FROM golang:1.23.5-alpine3.21 AS build

# Set metadata and maintainer
LABEL org.opencontainers.image.source="https://github.com/jfallot/shelly-blu-trv-exporter"
LABEL org.opencontainers.image.description="Prometheus Exporter for Shelly BLU TRV devices through Blu gateway"

# Install system dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy only go mod and sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies with caching
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with more robust flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -X 'main.Version=$(git describe --tags --always --dirty)'" \
    -o /bin/shelly-blu-trv-exporter

# Final stage with alpine
FROM alpine:3.21

# Create non-root user
RUN adduser -D -u 1000 appuser

# Copy the binary
COPY --from=build /bin/shelly-blu-trv-exporter /bin/shelly-blu-trv-exporter

# Copy config template
COPY --from=build /app/config.yaml.example /etc/shelly-blu-trv-exporter/config.yaml.example

# Use non-root user
USER appuser

# Default config volume
VOLUME ["/etc/shelly-blu-trv-exporter"]

# Expose metrics port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s \
  CMD wget -q -O- http://localhost:8080/metrics || exit 1

# Set entrypoint with config path
ENTRYPOINT ["/bin/shelly-blu-trv-exporter"]

# Default arguments can be overridden
CMD ["-c", "/etc/shelly-blu-trv-exporter/config.yaml"]