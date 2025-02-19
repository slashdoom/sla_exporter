package tcping

import (
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "sla_tcping_"

var (
	Result   *prometheus.GaugeVec
	Duration *prometheus.GaugeVec
)

func init() {
	l := []string{"target", "alias", "port", "test", "stat"}

	d := prometheus.GaugeOpts{Name: prefix + "result", Help: "Result of test"}
	Result = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "duration", Help: "Duration of test"}
	Duration = prometheus.NewGaugeVec(d, l)
}

func Register(registry *prometheus.Registry) {
	registry.MustRegister(Result)
	registry.MustRegister(Duration)
}

func Reset() {
	Result.Reset()
	Duration.Reset()
}

func Record(host string, alias string, port string, result TcpingResult) {
	l := prometheus.Labels{"target": host, "alias": alias, "port": port, "test": "tcping", "stat": "result"}
	completed := 0.00
	if result.Completed {
		completed = 1.00
	}
	Result.With(l).Set(completed)
	l = prometheus.Labels{"target": host, "alias": alias, "port": port, "test": "tcping", "stat": "duration"}
	Duration.With(l).Set(result.Duration.Seconds())
}
