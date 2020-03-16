package exporter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestScrapeMetrics(t *testing.T) {
	desc1 := `Desc{fqName: "linkquality_underlay_jitter_milliseconds", ` +
		`help: "Jitter in milliseconds for a specific type of service", ` +
		`constLabels: {}, variableLabels: [org edge link_name service]}`
	desc2 := `Desc{fqName: "linkquality_overlay_packet_loss", ` +
		`help: "Total packet loss for a specific type of service", ` +
		`constLabels: {}, variableLabels: [org edge service]}`

	m := NewMetricMapper("linkquality", []string{"org", "edge"})
	m.RegisterMetrics(
		&UnderlayQualityMetrics{},
		&OverlayQualityMetrics{},
	)

	AddMetricMapper(m)

	u := make([]*UnderlayQuality, 2)
	metrics0 := &UnderlayQualityMetrics{
		UnderlayJitterMilliseconds:        1.5,
		UnderlayPacketLoss:                10,
		UnderlayRoundtripTimeMilliseconds: 10.2,
	}
	metrics1 := &UnderlayQualityMetrics{
		UnderlayJitterMilliseconds:        2.3,
		UnderlayPacketLoss:                50,
		UnderlayRoundtripTimeMilliseconds: 40.6,
	}
	u[0] = &UnderlayQuality{LinkName: "wan0", VideoServiceMetrics: metrics0}
	u[1] = &UnderlayQuality{LinkName: "wan1", VoiceServiceMetrics: metrics1}
	metric2 := &OverlayQualityMetrics{
		OverlayJitterMilliseconds:        1.1,
		OverlayPacketLoss:                5,
		OverlayRoundtripTimeMilliseconds: 20.2,
	}
	o := &OverlayQuality{VoiceServiceMetrics: metric2}
	l := &LinkQuality{UnderlayQuality: u, OverlayQuality: o}

	ScrapeMetrics("linkquality", l, []string{"org", "edge"})

	metric, _ := m.getMetricVec("underlay_jitter_milliseconds").GetMetricWithLabelValues("org", "edge", "wan0", "video")
	if desc1 != metric.(prometheus.Untyped).Desc().String() {
		t.Fatalf("unexpected desc: %s\n", metric.(prometheus.Untyped).Desc().String())
	}
	valBits := math.Float64bits(1.5)
	if !isValEqual(metric, valBits) {
		t.Fatalf("wrong metric value: %+v\n", metric)
	}

	metric, _ = m.getMetricVec("underlay_jitter_milliseconds").GetMetricWithLabelValues("org", "edge", "wan1", "voice")
	if desc1 != metric.(prometheus.Untyped).Desc().String() {
		t.Fatalf("unexpected desc: %s\n", metric.(prometheus.Untyped).Desc().String())
	}
	valBits = math.Float64bits(2.3)
	if !isValEqual(metric, valBits) {
		t.Fatalf("wrong metric value: %+v\n", metric)
	}

	metric, _ = m.getMetricVec("overlay_packet_loss").GetMetricWithLabelValues("org", "edge", "voice")
	if desc2 != metric.(prometheus.Untyped).Desc().String() {
		t.Fatalf("unexpected desc: %s\n", metric.(prometheus.Untyped).Desc().String())
	}
	valBits = math.Float64bits(5)
	if !isValEqual(metric, valBits) {
		t.Fatalf("wrong metric value: %+v\n", metric)
	}
}

func isValEqual(metric prometheus.Metric, val uint64) bool {
	return strings.Contains(fmt.Sprintf("%+v\n", metric), strconv.FormatUint(val, 10))
}
