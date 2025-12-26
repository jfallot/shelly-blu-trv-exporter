# 🔌 Cheap Switch Exporter

Prometheus Exporter for low-cost network switches without SNMP support

## 📖 Overview

This Prometheus exporter retrieves Shelly BLU TRV statistics from Shelly BLU Gateway, enabling monitoring through a web-based interface.

## 🚀 Installation

### Prerequisites

- Go 1.23+
- Docker (optional)

### Direct Installation

1. Clone the repository
2. Download dependencies
```bash
go mod download
```

3. Copy configuration template
```bash
cp config.yaml.example config.yaml
```

4. Edit `config.yaml` with your switch details and parameters
5. Run the exporter
```bash
go run main.go
```

### Docker Deployment

```bash
# Build Docker image
docker build -t shelly-blu-trv-exporter .

# Run container
docker run -v "./config.yaml:/etc/shelly-blu-trv-exporter/config.yaml" -p 8080:8080 shelly-blu-trv-exporter
```

## 📝 Configuration

Create a `config.yaml` with the following structure:

```yaml
address: "192.168.1.1"           # IP or hostname of the switch
username: "admin"                # Web interface username
password: "password"             # Web interface password
timeout_seconds: 5               # Request timeout
```

## 📊 Exposed Metrics
blutrv_battery
blutrv_compone
blutrv_current_c
blutrv_flags
blutrv_identity
blutrv_info
blutrv_rssi
blutrv_target_c

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## 🐛 Issues

Report issues on the GitHub repository's issue tracker.
