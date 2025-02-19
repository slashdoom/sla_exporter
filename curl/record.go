package curl

import (
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "sla_curl_"

var (
	ResultGauge   *prometheus.GaugeVec
	DurationGauge *prometheus.GaugeVec

	dnsDurationGauge       *prometheus.GaugeVec
	connDurationGauge      *prometheus.GaugeVec
	tlsDurationGauge       *prometheus.GaugeVec
	ReqToRespDurationGauge *prometheus.GaugeVec

	statusCodeGauge *prometheus.GaugeVec
)

func init() {
	l := []string{"target", "alias", "test", "method", "stat"}

	d := prometheus.GaugeOpts{Name: prefix + "result", Help: "Result of test"}
	ResultGauge = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "duration", Help: "Duration of test"}
	DurationGauge = prometheus.NewGaugeVec(d, l)

	d = prometheus.GaugeOpts{Name: prefix + "dns_duration", Help: "Duration of DNS lookup"}
	dnsDurationGauge = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "conn_duration", Help: "Duration of connection dial"}
	connDurationGauge = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "tls_duration", Help: "Duration of TLS handshake"}
	tlsDurationGauge = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "req_to_resp_duration", Help: "Duration from request to response"}
	ReqToRespDurationGauge = prometheus.NewGaugeVec(d, l)

	d = prometheus.GaugeOpts{Name: prefix + "status_code", Help: "Status code"}
	statusCodeGauge = prometheus.NewGaugeVec(d, l)
}

func Register(registry *prometheus.Registry) {
	registry.MustRegister(ResultGauge)
	registry.MustRegister(DurationGauge)

	registry.MustRegister(dnsDurationGauge)
	registry.MustRegister(connDurationGauge)
	registry.MustRegister(tlsDurationGauge)
	registry.MustRegister(ReqToRespDurationGauge)

	registry.MustRegister(statusCodeGauge)
}

func Reset() {
	ResultGauge.Reset()
	DurationGauge.Reset()

	dnsDurationGauge.Reset()
	connDurationGauge.Reset()
	tlsDurationGauge.Reset()
	ReqToRespDurationGauge.Reset()

	statusCodeGauge.Reset()
}

func Record(url string, alias string, method string, result CurlResult) {
	l := prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "result"}
	completed := 0.00
	if result.Completed {
		completed = 1.00
	}
	ResultGauge.With(l).Set(completed)
	l = prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "duration"}
	DurationGauge.With(l).Set(result.Duration.Seconds())

	l = prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "dns_duration"}
	dnsDurationGauge.With(l).Set(float64(result.dnsDuration.Seconds()))
	l = prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "conn_duration"}
	connDurationGauge.With(l).Set(float64(result.connDuration.Seconds()))
	l = prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "tls_duration"}
	tlsDurationGauge.With(l).Set(float64(result.tlsDuration.Seconds()))
	l = prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "req_to_resp_duration"}
	ReqToRespDurationGauge.With(l).Set(float64(result.ReqToRespDuration.Seconds()))

	l = prometheus.Labels{"target": url, "alias": alias, "method": method, "test": "curl", "stat": "status_code"}
	statusCodeGauge.With(l).Set(float64(result.statusCode))
}
