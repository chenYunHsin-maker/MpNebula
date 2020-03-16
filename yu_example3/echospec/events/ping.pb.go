// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ping.proto

package events

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

//
type PingReport struct {
	DeviceSn  string                     `protobuf:"bytes,1,opt,name=device_sn,json=deviceSn" json:"device_sn"`
	Timestamp *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp"`
	TaskId    string                     `protobuf:"bytes,3,opt,name=task_id,json=taskId" json:"task_id"`
	Report    string                     `protobuf:"bytes,4,opt,name=report" json:"report"`
}

func (m *PingReport) Reset()                    { *m = PingReport{} }
func (m *PingReport) String() string            { return proto.CompactTextString(m) }
func (*PingReport) ProtoMessage()               {}
func (*PingReport) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{0} }

func (m *PingReport) GetDeviceSn() string {
	if m != nil {
		return m.DeviceSn
	}
	return ""
}

func (m *PingReport) GetTimestamp() *google_protobuf.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *PingReport) GetTaskId() string {
	if m != nil {
		return m.TaskId
	}
	return ""
}

func (m *PingReport) GetReport() string {
	if m != nil {
		return m.Report
	}
	return ""
}

func init() {
	proto.RegisterType((*PingReport)(nil), "events.PingReport")
}

func init() { proto.RegisterFile("ping.proto", fileDescriptor13) }

var fileDescriptor13 = []byte{
	// 175 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x8d, 0xb1, 0x0e, 0x82, 0x30,
	0x10, 0x86, 0x53, 0x35, 0x55, 0xce, 0xad, 0x83, 0x36, 0x38, 0x48, 0x9c, 0x98, 0x4a, 0xa2, 0x8b,
	0xaf, 0xe0, 0x66, 0xd0, 0x9d, 0x80, 0x9c, 0x4d, 0xa3, 0xb4, 0x0d, 0xad, 0x3c, 0x89, 0x0f, 0x6c,
	0x6c, 0x83, 0x8e, 0xf7, 0xe5, 0xbb, 0xff, 0x03, 0xb0, 0x4a, 0x4b, 0x61, 0x7b, 0xe3, 0x0d, 0xa3,
	0x38, 0xa0, 0xf6, 0x2e, 0xdd, 0x4a, 0x63, 0xe4, 0x13, 0x8b, 0x40, 0x9b, 0xd7, 0xbd, 0xf0, 0xaa,
	0x43, 0xe7, 0xeb, 0xce, 0x46, 0x71, 0xf7, 0x26, 0x00, 0x67, 0xa5, 0x65, 0x89, 0xd6, 0xf4, 0x9e,
	0x6d, 0x20, 0x69, 0x71, 0x50, 0x37, 0xac, 0x9c, 0xe6, 0x24, 0x23, 0x79, 0x52, 0x2e, 0x22, 0xb8,
	0x68, 0x76, 0x84, 0xe4, 0xf7, 0xce, 0x27, 0x19, 0xc9, 0x97, 0xfb, 0x54, 0xc4, 0x80, 0x18, 0x03,
	0xe2, 0x3a, 0x1a, 0xe5, 0x5f, 0x66, 0x6b, 0x98, 0xfb, 0xda, 0x3d, 0x2a, 0xd5, 0xf2, 0x69, 0x18,
	0xa5, 0xdf, 0xf3, 0xd4, 0xb2, 0x15, 0xd0, 0x3e, 0x94, 0xf9, 0x2c, 0xf2, 0x78, 0x35, 0x34, 0xec,
	0x1d, 0x3e, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8b, 0xf5, 0x7b, 0x2c, 0xd4, 0x00, 0x00, 0x00,
}
