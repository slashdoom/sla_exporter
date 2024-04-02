package dns

import (
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "sla_dns_"

var (
	Result   *prometheus.GaugeVec
	Duration *prometheus.GaugeVec
)


func init() {
	l := []string{"target", "server", "test", "stat"}

	d := prometheus.GaugeOpts{Name: prefix + "result", Help: "Result of DNS test"}
	Result = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "duration", Help: "Duration of DNS test in seconds"}
	Duration = prometheus.NewGaugeVec(d, l)
}


func Register(registry *prometheus.Registry) {
	registry.MustRegister(Result)
	registry.MustRegister(Duration)
}


func Record(host string, server string, result DnsTestResult) {
	l := prometheus.Labels{"target": host, "server": server, "test": "dns", "stat": "result"}
	completed := 0.00
	if result.Completed {
		completed = 1.00
	}
	Result.With(l).Set(completed)

	l = prometheus.Labels{"target": host, "server": server, "test": "dns", "stat": "duration"}
	Duration.With(l).Set(result.Duration.Seconds())
}
