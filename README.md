# sla_exporter

Prometheus exporter for SLA-like metrics.  Currently supported are: cURL (http/https), DNS, Ping (ICMP) and TCPing. 

Similar in a lot of ways to [Blackbox Exporter](https://github.com/prometheus/blackbox_exporter), and honestly most people will probably be better served by this.  This exporter was desgined to be a bit more basic as it was intended to be run on tele-operated forklifts and the operator stations for them.

# Flags
Name     | Description | Default
---------|-------------|---------
version | Print version information. |
web.listen-address | Address on which to expose metrics and web interface. | :9909
web.telemetry-path | Path under which to expose metrics. | /metrics
level | Set logging verbose level. | info
timeout | Global default test timeout. | 5
config.file | Path to config file (required). | ./config.yaml


# Install
```bash
go get -u github.com/slashdoom/sla_exporter
```

# Usage

## Binary
```bash
./sla_exporter -config.file=config.yml
```

## Docker
An example Docker compose file is included. It includes services for production, with 2 step build/deploy and Promethus container and one for a dev environment with hot reload.  These examples use the golang:alpine image.

### Production
```bash
docker compose up prometheus sla_exporter
```

### Development
```bash
docker compose up prometheus-dev sla_exporter-dev
```

## Config file
The exporter can be configured with a YAML based config file:

```yaml
---
level: debug
timeout: 5s

web:
  listen_address: 9909
  metrics_path: "/metrics"

curl_tests:
  - url: "http://example.com"
    timeout: 5s
  - url: "https://example.com"
    timeout: 5s
  - url: "https://slashdoom.com"
    timeout: 5s

dns_tests:
  - host: "example.com"
    timeout: 5s
  - host: "example.com"
    server: "9.9.9.9"
    timeout: 5s
  - host: "example.com"
    server: "8.8.8.8"
    timeout: 5s
  - host: "slashdoom.com"
    timeout: 5s

ping_tests:
  - host: "example.com"
    count: 3
    timeout: 5s
  - host: "slashdoom.com"
    count: 3
    timeout: 5s

tcping_tests:
  - host: "example.com"
    port: 80
    timeout: 5s
  - host: "example.com"
    port: 443
    timeout: 5s
  - host: "slashdoom.com"
    port: 443
    timeout: 5s
```

# Third Party Components
This software uses components of the following projects
* Prometheus Go client library (https://github.com/prometheus/client_golang)
* Prometheus Community pro-bing library (https://github.com/prometheus-community/pro-bing)
* Zap Logging library (https://github.com/uber-go/zap)

# License
(c) slashdoom (Patrick Ryon), 2024. Licensed under [MIT](LICENSE) license.

# Prometheus
See https://prometheus.io/
