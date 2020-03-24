package grpc

import (
	"sync"

	pbevent "yu_example3/events"
	pb "yu_example3/metrics"
)

// The structure here in this file is copied from grpc metrics/events and
// restructured as the required format by orchestrator and elasticsearch.

type newDNSAnswer struct {
	DeviceSn  string               `json:"device_sn"`
	Timestamp string               `json:"timestamp"`
	Entry     []*newDNSAnswerEntry `json:"entry"`
}

type newDNSAnswerEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Cls  string `json:"cls"`
	Addr string `json:"addr"`
	Fqdn string `json:"fqdn"`
	TTL  uint32 `json:"ttl"`
}

type newSystemLoad struct {
	DeviceSn  string             `json:"device_sn"`
	Timestamp string             `json:"timestamp"`
	CPU       newCPU             `json:"cpu"`
	Memory    *pb.MemoryMetrics  `json:"memory"`
	Network   *pb.NetworkMetrics `json:"network"`
}

type newCPU struct {
	TotalCore       int               `json:"total_core"`
	Utilization     map[string]uint32 `json:"utilization"`
	Utilization1Min map[string]uint32 `json:"utilization1min"`
	Utilization5Min map[string]uint32 `json:"utilization5min"`
}

type newDevicelog struct {
	Timestamp   int64                `json:"timestamp"`
	Severity    int32                `json:"severity"`
	Facility    int32                `json:"facility"`
	Category    string               `json:"category"`
	Srcip       string               `json:"srcip"`
	Dstip       string               `json:"dstip"`
	Ipproto     int32                `json:"ipproto"`
	Sport       int32                `json:"sport"`
	Dport       int32                `json:"dport"`
	DeviceSn    string               `json:"device_sn"`
	Logmessage  string               `json:"logmessage"`
	Note        string               `json:"note"`
	Username    string               `json:"username"`
	Srciface    string               `json:"srciface"`
	Dstiface    string               `json:"dstiface"`
	ProtoName   string               `json:"proto_name"`
	Devmac      []byte               `json:"devmac"`
	Count       uint32               `json:"count"`
	TrafficLog  *pbevent.TrafficLog  `json:"traffic_log"`
	IdpLog      *pbevent.IDPLog      `json:"idp_log"`
	FirewallLog *pbevent.FirewallLog `json:"firewall_log"`
	GeoSrc      *pbevent.GEOIPLog    `json:"geo_src"`
	GeoDst      *pbevent.GEOIPLog    `json:"geo_dst"`
}

// Define struct for history (reportTrafficInfo) monitoring data
// in one minutes range.
type reportTrafficInfoMap struct {
	m   map[string]newTrafficTuple
	mux sync.Mutex
}

type newPerLinkQuality struct {
	DeviceSn          string                    `json:"device_sn"`
	Timestamp         string                    `json:"timestamp"`
	PeerName          string                    `json:"peerName"`
	LinkName          string                    `json:"link_name"`
	QualityParameters *pb.LinkQualityParameters `json:"quality_parameters"`
	Failure           bool                      `json:"failure"`
}

type newPorts struct {
	DeviceSn  string                     `json:"device_sn"`
	Timestamp string                     `json:"timestamp"`
	Corp      string                     `json:"corp,omitempty"`
	Name      string                     `json:"name,omitempty"`
	SiteID    uint32                     `json:"siteId,omitempty"`
	PortName  string                     `json:"port_name"`
	State     pbevent.PortOperatingState `json:"state"`
	QueryKey  string                     `json:"query_key"`
}

type newARPEntry struct {
	DeviceSn   string `json:"device_sn"`
	Timestamp  string `json:"timestamp"`
	IPAddr     string `json:"ip_addr"`
	MacAddress string `json:"mac_address"`
	Device     string `json:"device"`
}

type newClientInfo struct {
	LoginName      string   `json:"login_name"`
	Hostname       string   `json:"hostname"`
	MacAddr        string   `json:"mac_addr"`
	Ipv4Addr       [][]byte `json:"ipv4_addr"`
	Ipv6Addr       [][]byte `json:"ipv6_addr"`
	DeviceSn       string   `json:"device_sn"`
	Timestamp      string   `json:"timestamp"`
	ProductName    string   `json:"product_name"`
	ProductVersion string   `json:"product_version"`
	ExtraOsInfo    string   `json:"extra_os_info"`
	DhcpPacket     []byte   `json:"dhcp_packet"`
}

type newTrafficTuple struct {
	ClientIP              []byte        `json:"client_ip"`
	ClientPort            uint32        `json:"client_port"`
	Proto                 pb.L4Protocol `json:"proto"`
	RemoteIP              []byte        `json:"remote_ip"`
	RemotePort            uint32        `json:"remote_port"`
	EstablishAt           int64         `json:"establish_at"`
	ReportAt              int64         `json:"report_at"`
	Application           uint32        `json:"application"`
	ClientHostname        string        `json:"client_hostname"`
	ClientUsername        string        `json:"client_username"`
	ClientOperatingSystem string        `json:"client_operating_system"`
	Qos                   pb.QosLevel   `json:"qos"`
	Status                pb.Status     `json:"status"`
	MacAddress            []byte        `json:"mac_address"`
	TxPackets             uint64        `json:"tx_packets"`
	TxOctets              uint64        `json:"tx_octets"`
	RxPackets             uint64        `json:"rx_packets"`
	RxOctets              uint64        `json:"rx_octets"`
	VtiTxPackets          uint64        `json:"vti_tx_packets"`
	VtiTxOctets           uint64        `json:"vti_tx_octets"`
	VtiRxPackets          uint64        `json:"vti_rx_packets"`
	VtiRxOctets           uint64        `json:"vti_rx_octets"`
}

func (m *newTrafficTuple) GetTxPackets() uint64 {
	if m != nil {
		return m.TxPackets
	}
	return 0
}

func (m *newTrafficTuple) GetTxOctets() uint64 {
	if m != nil {
		return m.TxOctets
	}
	return 0
}

func (m *newTrafficTuple) GetRxPackets() uint64 {
	if m != nil {
		return m.RxPackets
	}
	return 0
}

func (m *newTrafficTuple) GetRxOctets() uint64 {
	if m != nil {
		return m.RxOctets
	}
	return 0
}

func (m *newTrafficTuple) GetVtiTxPackets() uint64 {
	if m != nil {
		return m.VtiTxPackets
	}
	return 0
}

func (m *newTrafficTuple) GetVtiTxOctets() uint64 {
	if m != nil {
		return m.VtiTxOctets
	}
	return 0
}

func (m *newTrafficTuple) GetVtiRxPackets() uint64 {
	if m != nil {
		return m.VtiRxPackets
	}
	return 0
}

func (m *newTrafficTuple) GetVtiRxOctets() uint64 {
	if m != nil {
		return m.VtiRxOctets
	}
	return 0
}

func (m *newTrafficTuple) GetClientIP() []byte {
	if m != nil {
		return m.ClientIP
	}
	return nil
}

func (m *newTrafficTuple) GetClientPort() uint32 {
	if m != nil {
		return m.ClientPort
	}
	return 0
}

func (m *newTrafficTuple) GetProto() pb.L4Protocol {
	if m != nil {
		return m.Proto
	}
	return pb.L4Protocol_OTHER
}

func (m *newTrafficTuple) GetRemoteIP() []byte {
	if m != nil {
		return m.RemoteIP
	}
	return nil
}

func (m *newTrafficTuple) GetRemotePort() uint32 {
	if m != nil {
		return m.RemotePort
	}
	return 0
}

func (m *newTrafficTuple) GetEstablishAt() int64 {
	if m != nil {
		return m.EstablishAt
	}
	return 0
}

func (m *newTrafficTuple) GetReportAt() int64 {
	if m != nil {
		return m.ReportAt
	}
	return 0
}

func (m *newTrafficTuple) GetApplication() uint32 {
	if m != nil {
		return m.Application
	}
	return 0
}

func (m *newTrafficTuple) GetClientHostname() string {
	if m != nil {
		return m.ClientHostname
	}
	return ""
}

func (m *newTrafficTuple) GetClientUsername() string {
	if m != nil {
		return m.ClientUsername
	}
	return ""
}

func (m *newTrafficTuple) GetClientOperatingSystem() string {
	if m != nil {
		return m.ClientOperatingSystem
	}
	return ""
}

func (m *newTrafficTuple) GetQos() pb.QosLevel {
	if m != nil {
		return m.Qos
	}
	return pb.QosLevel_CONTROL
}

func (m *newTrafficTuple) GetStatus() pb.Status {
	if m != nil {
		return m.Status
	}
	return pb.Status_CREATED
}

func (m *newTrafficTuple) GetMacAddress() []byte {
	if m != nil {
		return m.MacAddress
	}
	return nil
}
