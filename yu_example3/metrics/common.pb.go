// Code generated by protoc-gen-go. DO NOT EDIT.
// source: common.proto

/*
Package metrics is a generated protocol buffer package.

It is generated from these files:
	common.proto
	cubs.proto
	cubs_interface_livemon.proto
	cubs_linkquality.proto
	cubs_periodical_report.proto
	cubs_system.proto
	cubs_vpn_traffic.proto
	turboproxy.proto

It has these top-level messages:
	Empty
	LiveReport
	InterfaceStat
	InterfaceMetrics
	L4ProtocolMetrics
	LinkQualityParameters
	PerLinkQuality
	PeerLinkQuality
	LinkQuality
	TrafficMetrics
	TrafficVolume
	TrafficTuple
	AccumulatedTraffic
	DeltaTraffic
	FlowTrafficInfo
	CPUMetrics
	CPUCore
	CPULoad
	MemoryMetrics
	NetworkMetrics
	SystemLoad
	VPNTrafficInfo
	VPNConnInfo
	PacketAndOctetMetrics
	DropMetrics
	SCTPStat
	SCTPInputMetrics
	SCTPOutputMetrics
	SCTPCongestionMetrics
	SCTPDropMetrics
	SCTPTimeoutMetrics
	SCTPOtherMetrics
*/
package metrics

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Empty is the response to any ReportXXX rpc.
type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func init() {
	proto.RegisterType((*Empty)(nil), "metrics.Empty")
}

func init() { proto.RegisterFile("common.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 61 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0xce, 0xcf, 0xcd,
	0xcd, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xcf, 0x4d, 0x2d, 0x29, 0xca, 0x4c,
	0x2e, 0x56, 0x62, 0xe7, 0x62, 0x75, 0xcd, 0x2d, 0x28, 0xa9, 0x4c, 0x62, 0x03, 0x4b, 0x18, 0x03,
	0x02, 0x00, 0x00, 0xff, 0xff, 0x88, 0x10, 0xf7, 0xa8, 0x28, 0x00, 0x00, 0x00,
}
