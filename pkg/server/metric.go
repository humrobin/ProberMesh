package server

import "github.com/prometheus/client_golang/prometheus"

var (
	// icmp
	icmpProberFailedGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_icmp_failed",
		Help: "icmp prober failed times",
	}, []string{"source_region", "target_region"})

	icmpProberDurationGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_icmp_duration_seconds",
		Help: "icmp prober duration by phase",
	}, []string{"phase", "source_region", "target_region"})

	// http
	httpProberFailedGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_http_failed",
		Help: "http prober failed times",
	}, []string{"source_region", "target_addr"})

	httpProberDurationGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_http_duration_seconds",
		Help: "http prober duration by phase",
	}, []string{"phase", "source_region", "target_addr"})
)

func init() {
	prometheus.MustRegister(icmpProberFailedGaugeVec)
	prometheus.MustRegister(icmpProberDurationGaugeVec)
	prometheus.MustRegister(httpProberFailedGaugeVec)
	prometheus.MustRegister(httpProberDurationGaugeVec)
}
