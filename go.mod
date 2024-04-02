module sla_exporter

go 1.22.1

require (
	example.org/config v0.0.0
	example.org/curl v0.0.0
	example.org/dns v0.0.0
	example.org/logger v0.0.0
	example.org/ping v0.0.0
	example.org/tcping v0.0.0
	github.com/prometheus/client_golang v1.19.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus-community/pro-bing v0.4.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	example.org/config => ./config
	example.org/curl => ./curl
	example.org/dns => ./dns
	example.org/logger => ./logger
	example.org/ping => ./ping
	example.org/tcping => ./tcping
)
