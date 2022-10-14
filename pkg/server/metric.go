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

	icmpProberDurationHistogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "prober_icmp_duration_seconds_total",
		Help: "icmp prober duration histogram by phase",
		// 5e-05,0.0001,0.0002,0.0004,0.0008,0.0016,0.0032,0.0064,0.0128,0.0256,0.0512,0.1024,0.2048,0.4096,0.8192,1.6384,3.2768,6.5536,13.1072
		Buckets: prometheus.ExponentialBuckets(0.00005, 2, 19),
	}, []string{"source_region", "target_region"})

	// http
	httpProberFailedGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_http_failed",
		Help: "http prober failed times",
	}, []string{"source_region", "target_addr"})

	httpProberDurationGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_http_duration_seconds",
		Help: "http prober duration by phase",
	}, []string{"phase", "source_region", "target_addr"})

	// healthCheck
	agentHealthCheckGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "prober_agent_is_alive",
		Help: "proberMesh agent is alive",
	}, []string{"region", "ip"})
)

func init() {
	prometheus.MustRegister(icmpProberFailedGaugeVec)
	prometheus.MustRegister(icmpProberDurationGaugeVec)
	prometheus.MustRegister(icmpProberDurationHistogramVec)
	prometheus.MustRegister(httpProberFailedGaugeVec)
	prometheus.MustRegister(httpProberDurationGaugeVec)
	prometheus.MustRegister(agentHealthCheckGaugeVec)
}
