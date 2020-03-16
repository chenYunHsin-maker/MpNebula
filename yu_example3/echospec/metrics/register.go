package metrics

import "sdn.io/sdwan/pkg/monitorproxy/exporter"

// MustRegister registers all metrics defined in this package and
// panics if any error occurs.
// Metrics are grouped by subsystem or module that each group is
// corresponding to a MetricMapper.
func MustRegister() {
	labelNames := []string{"org", "device"}

	m := exporter.NewMetricMapper("turboproxy", labelNames)
	m.RegisterMetrics(
		&SCTPInputMetrics{},
		&SCTPOutputMetrics{},
		&SCTPCongestionMetrics{},
		&SCTPDropMetrics{},
		&SCTPTimeoutMetrics{},
		&SCTPOtherMetrics{},
	)
	exporter.AddMetricMapper(m)

	m = exporter.NewMetricMapper("interface_livemon", labelNames)
	m.RegisterMetrics(
		&InterfaceMetrics{},
		&L4ProtocolMetrics{},
	)
	exporter.AddMetricMapper(m)

	m = exporter.NewMetricMapper("vpn_traffic", labelNames)
	m.RegisterMetrics(
		&PacketAndOctetMetrics{},
		&DropMetrics{},
	)

	exporter.AddMetricMapper(m)
}
