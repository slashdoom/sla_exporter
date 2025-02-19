package ping

import (
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "sla_ping_"

var (
	Result   *prometheus.GaugeVec
	Duration *prometheus.GaugeVec

	PacketsRecv    *prometheus.GaugeVec
	PacketsSent    *prometheus.GaugeVec
	PacketsRecvDup *prometheus.GaugeVec
	PacketLoss     *prometheus.GaugeVec

	MinRtt    *prometheus.GaugeVec
	MaxRtt    *prometheus.GaugeVec
	AvgRtt    *prometheus.GaugeVec
	StdDevRtt *prometheus.GaugeVec
)

func init() {
	l := []string{"target", "alias", "test", "stat"}

	d := prometheus.GaugeOpts{Name: prefix + "result", Help: "Result of test"}
	Result = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "duration", Help: "Duration of test"}
	Duration = prometheus.NewGaugeVec(d, l)

	d = prometheus.GaugeOpts{Name: prefix + "packets_recv", Help: "Number of packets received"}
	PacketsRecv = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "packets_sent", Help: "Number of packets sent"}
	PacketsSent = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "packets_recv_dup", Help: "Number of duplicate packets received"}
	PacketsRecvDup = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "packet_loss", Help: "Percentage of packets lost"}
	PacketLoss = prometheus.NewGaugeVec(d, l)

	d = prometheus.GaugeOpts{Name: prefix + "min_rtt", Help: "Minimum round-trip time"}
	MinRtt = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "max_rrt", Help: "Maximum round-trip time"}
	MaxRtt = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "avg_rtt", Help: "Average round-trip time"}
	AvgRtt = prometheus.NewGaugeVec(d, l)
	d = prometheus.GaugeOpts{Name: prefix + "std_dev_rtt", Help: "Standard deviation of the round-trip"}
	StdDevRtt = prometheus.NewGaugeVec(d, l)
}

func Register(registry *prometheus.Registry) {
	registry.MustRegister(Result)
	registry.MustRegister(Duration)

	registry.MustRegister(PacketsRecv)
	registry.MustRegister(PacketsSent)
	registry.MustRegister(PacketsRecvDup)
	registry.MustRegister(PacketLoss)

	registry.MustRegister(MinRtt)
	registry.MustRegister(MaxRtt)
	registry.MustRegister(AvgRtt)
	registry.MustRegister(StdDevRtt)
}

func Reset() {
	Result.Reset()
	Duration.Reset()

	PacketsRecv.Reset()
	PacketsSent.Reset()
	PacketsRecvDup.Reset()
	PacketLoss.Reset()

	MinRtt.Reset()
	MaxRtt.Reset()
	AvgRtt.Reset()
	StdDevRtt.Reset()
}

func Record(host string, alias string, result PingResult) {
	l := prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "result"}
	completed := 0.00
	if result.Completed {
		completed = 1.00
	}
	Result.With(l).Set(completed)

	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "duration"}
	Duration.With(l).Set(result.Duration.Seconds())

	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "packets_recv"}
	PacketsRecv.With(l).Set(float64(result.PacketsRecv))
	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "packets_sent"}
	PacketsSent.With(l).Set(float64(result.PacketsSent))
	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "packets_recv_dup"}
	PacketsRecvDup.With(l).Set(float64(result.PacketsRecvDup))
	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "packet_loss"}
	PacketLoss.With(l).Set(float64(result.PacketLoss))

	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "min_rtt"}
	MinRtt.With(l).Set(result.MinRtt.Seconds())
	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "max_rtt"}
	MaxRtt.With(l).Set(result.MaxRtt.Seconds())
	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "avg_rtt"}
	AvgRtt.With(l).Set(result.AvgRtt.Seconds())
	l = prometheus.Labels{"target": host, "alias": alias, "test": "ping", "stat": "std_dev_rtt"}
	StdDevRtt.With(l).Set(result.StdDevRtt.Seconds())
}
