package exporter

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type LinkQuality struct {
	UnderlayQuality []*UnderlayQuality `protobuf:"bytes,1,rep,name=underlay_quality,json=underlayQuality" json:"underlay_quality,omitempty"`
	OverlayQuality  *OverlayQuality    `protobuf:"bytes,2,opt,name=overlay_quality,json=overlayQuality" json:"overlay_quality,omitempty"`
}

type UnderlayQuality struct {
	// @inject_tag: label:"link_name=?"
	LinkName string `protobuf:"bytes,1,opt,name=link_name,json=linkName" json:"link_name,omitempty" label:"link_name=?"`
	// @inject_tag: label:"service=transactional"
	TransactionalServiceMetrics *UnderlayQualityMetrics `protobuf:"bytes,2,opt,name=transactional_service_metrics,json=transactionalServiceMetrics" json:"transactional_service_metrics,omitempty" label:"service=transactional"`
	// @inject_tag: label:"service=video"
	VideoServiceMetrics *UnderlayQualityMetrics `protobuf:"bytes,3,opt,name=video_service_metrics,json=videoServiceMetrics" json:"video_service_metrics,omitempty" label:"service=video"`
	// @inject_tag: label:"service=voice"
	VoiceServiceMetrics *UnderlayQualityMetrics `protobuf:"bytes,4,opt,name=voice_service_metrics,json=voiceServiceMetrics" json:"voice_service_metrics,omitempty" label:"service=voice"`
}

type OverlayQuality struct {
	// @inject_tag: label:"service=transactional"
	TransactionalServiceMetrics *OverlayQualityMetrics `protobuf:"bytes,1,opt,name=transactional_service_metrics,json=transactionalServiceMetrics" json:"transactional_service_metrics,omitempty" label:"service=transactional"`
	// @inject_tag: label:"service=video"
	VideoServiceMetrics *OverlayQualityMetrics `protobuf:"bytes,2,opt,name=video_service_metrics,json=videoServiceMetrics" json:"video_service_metrics,omitempty" label:"service=video"`
	// @inject_tag: label:"service=voice"
	VoiceServiceMetrics *OverlayQualityMetrics `protobuf:"bytes,3,opt,name=voice_service_metrics,json=voiceServiceMetrics" json:"voice_service_metrics,omitempty" label:"service=voice"`
}

type UnderlayQualityMetrics struct {
	// @inject_tag: labelnames:"link_name,service", help:"Jitter in milliseconds for a specific type of service"
	UnderlayJitterMilliseconds float64 `protobuf:"fixed64,1,opt,name=underlay_jitter_milliseconds,json=underlayJitterMilliseconds" json:"underlay_jitter_milliseconds,omitempty" labelnames:"link_name,service" help:"Jitter in milliseconds for a specific type of service"`
	// @inject_tag: help:"Total packet loss for a specific type of service"
	UnderlayPacketLoss uint64 `protobuf:"varint,2,opt,name=underlay_packet_loss,json=underlayPacketLoss" json:"underlay_packet_loss,omitempty" help:"Total packet loss for a specific type of service"`
	// @inject_tag: help:"Roundtrip time in milliseconds for a specific type of service"
	UnderlayRoundtripTimeMilliseconds float64 `protobuf:"fixed64,3,opt,name=underlay_roundtrip_time_milliseconds,json=underlayRoundtripTimeMilliseconds" json:"underlay_roundtrip_time_milliseconds,omitempty" help:"Roundtrip time in milliseconds for a specific type of service"`
}

type OverlayQualityMetrics struct {
	// @inject_tag: labelnames:"service", help:"Jitter in milliseconds for a specific type of service"
	OverlayJitterMilliseconds float64 `protobuf:"fixed64,1,opt,name=overlay_jitter_milliseconds,json=overlayJitterMilliseconds" json:"overlay_jitter_milliseconds,omitempty" labelnames:"service" help:"Jitter in milliseconds for a specific type of service"`
	// @inject_tag: help:"Total packet loss for a specific type of service"
	OverlayPacketLoss uint64 `protobuf:"varint,2,opt,name=overlay_packet_loss,json=overlayPacketLoss" json:"overlay_packet_loss,omitempty" help:"Total packet loss for a specific type of service"`
	// @inject_tag: help:"Roundtrip time in milliseconds for a specific type of service"
	OverlayRoundtripTimeMilliseconds float64 `protobuf:"fixed64,3,opt,name=overlay_roundtrip_time_milliseconds,json=overlayRoundtripTimeMilliseconds" json:"overlay_roundtrip_time_milliseconds,omitempty" help:"Roundtrip time in milliseconds for a specific type of service"`
}

func TestRegisterMetrics(t *testing.T) {
	m := NewMetricMapper("edge_linkquality", []string{"edge"})
	m.RegisterMetrics(
		&UnderlayQualityMetrics{},
	)

	jitter := m.getMetricVec("underlay_jitter_milliseconds")
	jitter.WithLabelValues("hongkong", "wan0", "video").(prometheus.Untyped).Set(1.5)
	packetLoss := m.getMetricVec("underlay_packet_loss")
	packetLoss.WithLabelValues("hongkong", "wan1", "voice").(prometheus.Untyped).Set(10)
	rtt := m.getMetricVec("underlay_roundtrip_time_milliseconds")
	rtt.WithLabelValues("hongkong", "wan1", "voice").(prometheus.Untyped).Set(10.2)
}
