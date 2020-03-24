package grpc

import (
	"encoding/json"
	//"errors"
	"fmt"
	"io"
	"net"
	"regexp"

	//"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/robfig/cron"

	//"sdn.io/sdwan/cmd/cubs/monitorproxy/apiclient"
	pbevent "yu_example3/events"
	pb "yu_example3/metrics"
)

var cronJob *cron.Cron

// Initilize an empty map for history monitoring data
var rti = reportTrafficInfoMap{m: make(map[string]newTrafficTuple), mux: sync.Mutex{}}

func (s *reportTrafficInfoMap) deleteMapObject(key string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.m, key)
}

func (s *reportTrafficInfoMap) addMapObject(key string, tv *newTrafficTuple) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.m[key] = *tv
}

func (s *reportTrafficInfoMap) hasMapObject(key string) (newTrafficTuple, bool) {
	var q newTrafficTuple
	s.mux.Lock()
	defer s.mux.Unlock()
	if val, ok := s.m[key]; ok {
		return val, true
	}
	return q, false
}

// deleteOldTrafficInfo will print reportTrafficInfoMap map length and delete timeout
// (>3600s) entry in the map every half an hour.
// Please note this function do NOT print reportTrafficInfo anymore.
func (s *reportTrafficInfoMap) deleteOldTrafficInfo() {
	s.mux.Lock()
	defer s.mux.Unlock()
	glog.Infof("The length for reporttrafficinfo map is %v", len(s.m))
	for k, v := range s.m {
		t := v.GetReportAt()
		if t != 0 {
			// this connection is not updated more than 3600 seconds, remove it from traffic info
			if int64(time.Now().Unix())-t >= 3600 {
				glog.V(10).Infof("Timeout for connection %v and deleted from traffic info mapping", v)
				delete(s.m, k)
				continue
			}
		}
	}
}

// deleteOldTrafficInfo create cron job for running deleteOldTrafficInfo
// every half an hour.
func deleteOldTrafficInfo() {
	glog.Info("Start running deleteOldTrafficInfo for reportTrafficInfo every half an hour")
	cronJob = cron.New()
	cronJob.AddFunc("0 */30 * * * *", rti.deleteOldTrafficInfo)
	cronJob.Start()
}

// StopCronJob stop the cronjob in program graceful exit.
func StopCronJob() {
	cronJob.Stop()
}

func getDeltaMetrics(new uint64, old uint64) uint64 {
	var delta uint64
	if new < old {
		// current counter is smaller than previous counter, send warning and set delta to 0
		delta = 0
		glog.Warningf("Counter decrease error detected, new counter: %+v, old counter: %+v", new, old)
		return delta
	}

	delta = new - old
	return delta
}

func printFlowTrafficInfo(m *pb.FlowTrafficInfo) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var deltaTxPackets, deltaRxPackets, deltaTxOctets, deltaRxOctets uint64
	var deltaVtiTxPackets, deltaVtiRxPackets, deltaVtiTxOctets, deltaVtiRxOctets uint64
	var sumTxPackets, sumRxPackets, sumTxOctets, sumRxOctets uint64
	var sumVtiTxPackets, sumVtiRxPackets, sumVtiTxOctets, sumVtiRxOctets uint64
	var inf string
	var isVTI, hasWAN, hasVTI bool
	var deltaTrafficTuple = make(map[string]*pb.TrafficTuple, len(m.GetDeltaTraffic().GetTrafficTuple()))

	// For firmware does not include SN, insert an none SN
	sn := getSN(m.GetDeviceSn())

	// For new firmware has RPC FlowTrafficInfo, store the delta in the hash map.
	for _, j := range m.GetDeltaTraffic().GetTrafficTuple() {
		deltaTrafficTuple[composeNetTupleWOMagic(j)] = j
	}

	// For history traffic monitoring, after discussing with xiaodong and Zyxel,
	// We have agreed to collect a reporttrafficinfo snapshots with maps which key
	// is "sn+tuple", and value is a struct of TxOctets, TxPackets,
	// RxOctets and RxPackets. The reporttrafficinfo snapshot will be collected based
	// on the metric time.
	for _, j := range m.GetAccumulatedTraffic().GetTrafficTuple() {
		// compose key string using five tuple with SN
		key := sn + "_" + composeNetTuple(j)

		// get time in RFC3339 format
		t3339 := time.Unix(j.GetReportAt(), 0).Format(time.RFC3339)
		sumTxPackets, sumRxPackets, sumTxOctets, sumRxOctets = 0, 0, 0, 0
		sumVtiTxPackets, sumVtiRxPackets, sumVtiTxOctets, sumVtiRxOctets = 0, 0, 0, 0
		deltaTxPackets, deltaRxPackets, deltaTxOctets, deltaRxOctets = 0, 0, 0, 0
		deltaVtiTxPackets, deltaVtiRxPackets, deltaVtiTxOctets, deltaVtiRxOctets = 0, 0, 0, 0
		hasWAN, hasVTI = false, false

		for _, k := range j.TrafficVolume {
			// get traffic volume and interface in the volume
			tvMetrics := k.GetMetrics()
			inf := k.GetInterface()

			if !strings.Contains(inf, "vti") {
				// if this interface is not VTI interface and it is LAN interface, do nothing.
				if strings.Contains(inf, "WAN") || strings.Contains(inf, "Cel") {
					// if this interface is not VTI interface and contain WAN or Cel, add metrics to the sum
					sumTxPackets = sumTxPackets + tvMetrics.GetTxPackets()
					sumRxPackets = sumRxPackets + tvMetrics.GetRxPackets()
					sumTxOctets = sumTxOctets + tvMetrics.GetTxOctets()
					sumRxOctets = sumRxOctets + tvMetrics.GetRxOctets()
					hasWAN = true
				}
			} else {
				//if this interface is VTI interface, add metrics to VTI sum
				sumVtiTxPackets = sumVtiTxPackets + tvMetrics.GetTxPackets()
				sumVtiRxPackets = sumVtiRxPackets + tvMetrics.GetRxPackets()
				sumVtiTxOctets = sumVtiTxOctets + tvMetrics.GetTxOctets()
				sumVtiRxOctets = sumVtiRxOctets + tvMetrics.GetRxOctets()
				hasVTI = true
			}
		}

		// Match the same flow from accumulate traffic to get delta traffic in the
		// FlowTrafficInfo. This for loop calculate the sum delta traffic of traffic
		// volume in the traffic tuple.
		for _, k := range deltaTrafficTuple[composeNetTupleWOMagic(j)].GetTrafficVolume() {
			// get traffic volume and interface in the volume
			tvMetrics := k.GetMetrics()
			inf := k.GetInterface()

			if !strings.Contains(inf, "vti") {
				// if this interface is not VTI interface and it is LAN interface, do nothing.
				if strings.Contains(inf, "WAN") || strings.Contains(inf, "Cel") {
					// if this interface is not VTI interface and contain WAN or Cel, add metrics to the sum
					deltaTxPackets = deltaTxPackets + tvMetrics.GetTxPackets()
					deltaRxPackets = deltaRxPackets + tvMetrics.GetRxPackets()
					deltaTxOctets = deltaTxOctets + tvMetrics.GetTxOctets()
					deltaRxOctets = deltaRxOctets + tvMetrics.GetRxOctets()
				}
			} else {
				//if this interface is VTI interface, add metrics to VTI sum
				deltaVtiTxPackets = deltaVtiTxPackets + tvMetrics.GetTxPackets()
				deltaVtiRxPackets = deltaVtiRxPackets + tvMetrics.GetRxPackets()
				deltaVtiTxOctets = deltaVtiTxOctets + tvMetrics.GetTxOctets()
				deltaVtiRxOctets = deltaVtiRxOctets + tvMetrics.GetRxOctets()
			}
		}

		var mac net.HardwareAddr = j.GetMacAddress()
		// if mac address field is empty, use "00:00:00:00:00:00" instead.
		if mac == nil {
			mac, _ = net.ParseMAC("00:00:00:00:00:00")
		}

		// Each traffic tuple will print two metrics log after discussing with Glenn and Yo.
		// This will reduce metrics log for each traffic tuple to 2, one is VTI, another is
		// non-VTI interface. Previously we could have one metrics log for each traffic volume
		// in traffic tuple.

		// The interface in the metrics log is no longer used but we will keep this field here
		// to maintain database backward compatibility. Also there is no need to update logstash and
		// elasticsearch configuration. Since orch use WAN1 and WAN2 to filter traffic in source
		// application and destination tab in monitor page, we will use WAN1 here to match the
		// orch elasticsearch search.
		inf = "WAN1"
		isVTI = false
		// Only print the log if has traffic volume on WAN or VTI interface
		if hasWAN {
			metrics.Printf("ReportTrafficInfo: Time: %v SN: %s "+
				"ClientIP: %v ClientPort: %v RemoteIP: %v RemotePort: %v Protocol: %v Interface: %v "+
				"Application: %v ClientHostname: %v ClientUsername: %v ClientOperatingSystem: %v Qos: %v "+
				"Status: %v MacAddress: %v isVTI: %v Magic: %v IcmpID: %v "+
				"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v "+
				"DeltaTxPackets: %v DeltaRxPackets: %v DeltaTxOctets: %v DeltaRxOctets: %v",
				t3339, sn, net.IP(j.GetClientIp()), j.GetClientPort(), net.IP(j.GetRemoteIp()), j.GetRemotePort(),
				j.GetProto(), inf, j.GetApplication(), j.GetClientHostname(), j.GetClientUsername(),
				j.GetClientOperatingSystem(), j.GetQos(), j.GetStatus(), mac.String(), isVTI, j.GetMagic(), j.GetIcmpid(),
				sumTxOctets, sumTxPackets, sumRxOctets, sumRxPackets,
				deltaTxPackets, deltaRxPackets, deltaTxOctets, deltaRxOctets)
		}

		inf = "WAN1"
		isVTI = true
		if hasVTI {
			metrics.Printf("ReportTrafficInfo: Time: %v SN: %s "+
				"ClientIP: %v ClientPort: %v RemoteIP: %v RemotePort: %v Protocol: %v Interface: %v "+
				"Application: %v ClientHostname: %v ClientUsername: %v ClientOperatingSystem: %v Qos: %v "+
				"Status: %v MacAddress: %v isVTI: %v Magic: %v IcmpID: %v "+
				"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v "+
				"DeltaTxPackets: %v DeltaRxPackets: %v DeltaTxOctets: %v DeltaRxOctets: %v",
				t3339, sn, net.IP(j.GetClientIp()), j.GetClientPort(), net.IP(j.GetRemoteIp()), j.GetRemotePort(),
				j.GetProto(), inf, j.GetApplication(), j.GetClientHostname(), j.GetClientUsername(),
				j.GetClientOperatingSystem(), j.GetQos(), j.GetStatus(), mac.String(), isVTI, j.GetMagic(), j.GetIcmpid(),
				sumVtiTxOctets, sumVtiTxPackets, sumVtiRxOctets, sumVtiRxPackets,
				deltaVtiTxPackets, deltaVtiRxPackets, deltaVtiTxOctets, deltaVtiRxOctets)
		}
		// the status is destroy, remove this connection from traffic info map
		if j.GetStatus() == pb.Status_DESTROY {
			glog.V(10).Infof("Destroy status received, deleted this connection %v from traffic info mapping", j)
			rti.deleteMapObject(key)
			continue
		}

		// add this traffic volume to hash map object
		k := copyNewTrafficTuple(j, sumTxPackets, sumRxPackets, sumTxOctets, sumRxOctets, sumVtiTxPackets, sumVtiRxPackets, sumVtiTxOctets, sumVtiRxOctets)
		rti.addMapObject(key, &k)
	}
	return nil
}

func printReportTrafficInfo(m *pb.AccumulatedTraffic) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var deltaTxPackets, deltaRxPackets, deltaTxOctets, deltaRxOctets uint64
	var deltaVtiTxPackets, deltaVtiRxPackets, deltaVtiTxOctets, deltaVtiRxOctets uint64
	var sumTxPackets, sumRxPackets, sumTxOctets, sumRxOctets uint64
	var sumVtiTxPackets, sumVtiRxPackets, sumVtiTxOctets, sumVtiRxOctets uint64
	var inf string
	var isVTI, hasWAN, hasVTI bool

	// For firmware does not include SN, insert an none SN
	sn := getSN(m.GetDeviceSn())

	// For history traffic monitoring, after discussing with xiaodong and Zyxel,
	// We have agreed to collect a reporttrafficinfo snapshots with maps which key
	// is "sn+tuple", and value is a struct of TxOctets, TxPackets,
	// RxOctets and RxPackets. The reporttrafficinfo snapshot will be collected based
	// on the metric time.
	// 2019-07-10 Add deltaTxPackets, deltaRxPackets, deltaTxOctets, deltaRxOctets
	// fields to the reporttrafficinfo metrics as discussed with Yo and xiaodong.
	// 2019-10-22 Refactor map structure and add all traffic volume to two sum metrics
	// one is VTI, another is non-VTI. The LAN interface will by passed from sum.
	for _, j := range m.TrafficTuple {
		// compose key string using five tuple with SN
		key := sn + "_" + composeNetTuple(j)

		// get time in RFC3339 format
		t3339 := time.Unix(j.GetReportAt(), 0).Format(time.RFC3339)

		sumTxPackets, sumRxPackets, sumTxOctets, sumRxOctets = 0, 0, 0, 0
		sumVtiTxPackets, sumVtiRxPackets, sumVtiTxOctets, sumVtiRxOctets = 0, 0, 0, 0
		hasWAN, hasVTI = false, false

		for _, k := range j.TrafficVolume {
			// get traffic volume and interface in the volume
			tvMetrics := k.GetMetrics()
			inf = k.GetInterface()

			if !strings.Contains(inf, "vti") {
				// if this interface is not VTI interface and it is LAN interface, do nothing.
				if strings.Contains(inf, "WAN") || strings.Contains(inf, "Cel") {
					// if this interface is not VTI interface and contain WAN or Cel, add metrics to the sum
					sumTxPackets = sumTxPackets + tvMetrics.GetTxPackets()
					sumRxPackets = sumRxPackets + tvMetrics.GetRxPackets()
					sumTxOctets = sumTxOctets + tvMetrics.GetTxOctets()
					sumRxOctets = sumRxOctets + tvMetrics.GetRxOctets()
					hasWAN = true
				}
			} else {
				//if this interface is VTI interface, add metrics to VTI sum
				sumVtiTxPackets = sumVtiTxPackets + tvMetrics.GetTxPackets()
				sumVtiRxPackets = sumVtiRxPackets + tvMetrics.GetRxPackets()
				sumVtiTxOctets = sumVtiTxOctets + tvMetrics.GetTxOctets()
				sumVtiRxOctets = sumVtiRxOctets + tvMetrics.GetRxOctets()
				hasVTI = true
			}
		}

		// check if the report traffic info map already has the key
		if v, ok := rti.hasMapObject(key); !ok {
			// traffic info map does not has this key and set delta to initial packet/octet data
			if j.GetStatus() == pb.Status_DESTROY {
				// there is no create before destroy, bypass this traffic tuple
				glog.V(10).Infof("received a destroy status in traffic tuple before create.")
				continue
			}
			if j.GetStatus() == pb.Status_UPDATE {
				// Dicussed with Rick, if the first status is update, use 0 instead of rx/tx as delta. Because the
				// accumulated coutner could be very large, if we use that as delta there will be a peak delta counter.
				deltaTxPackets = 0
				deltaRxPackets = 0
				deltaTxOctets = 0
				deltaRxOctets = 0
				deltaVtiTxPackets = 0
				deltaVtiRxPackets = 0
				deltaVtiTxOctets = 0
				deltaVtiRxOctets = 0
			} else if j.GetStatus() == pb.Status_CREATED {
				// use first create as delta
				deltaTxPackets = sumTxPackets
				deltaRxPackets = sumRxPackets
				deltaTxOctets = sumTxOctets
				deltaRxOctets = sumRxOctets
				deltaVtiTxPackets = sumVtiTxPackets
				deltaVtiRxPackets = sumVtiRxPackets
				deltaVtiTxOctets = sumVtiTxOctets
				deltaVtiRxOctets = sumVtiRxOctets
			}
		} else {
			// traffic info map already has this key and calculate the delta of metrics
			if j.GetStatus() == pb.Status_CREATED {
				// traffic info map already has this key however the new received status is create,
				// which means the destroy is not correctly received or missed.
				// So set delta to initial create data.
				deltaTxPackets = sumTxPackets
				deltaRxPackets = sumRxPackets
				deltaTxOctets = sumTxOctets
				deltaRxOctets = sumRxOctets
				deltaVtiTxPackets = sumVtiTxPackets
				deltaVtiRxPackets = sumVtiRxPackets
				deltaVtiTxOctets = sumVtiTxOctets
				deltaVtiRxOctets = sumVtiRxOctets
			} else {
				// retrice old metrics in the map
				oldTxPackets := v.GetTxPackets()
				oldRxPackets := v.GetRxPackets()
				oldTxOctets := v.GetTxOctets()
				oldRxOctets := v.GetRxOctets()
				oldVtiTxPackets := v.GetVtiTxPackets()
				oldVtiRxPackets := v.GetVtiRxPackets()
				oldVtiTxOctets := v.GetVtiTxOctets()
				oldVtiRxOctets := v.GetVtiRxOctets()

				//check if current counter is smaller than previous counter and calculate the delta
				var max uint64
				max = 180000000000
				deltaTxPackets = getDeltaMetrics(sumTxPackets, oldTxPackets)
				deltaVtiTxPackets = getDeltaMetrics(sumVtiTxPackets, oldVtiTxPackets)
				deltaRxPackets = getDeltaMetrics(sumRxPackets, oldRxPackets)
				deltaVtiRxPackets = getDeltaMetrics(sumVtiRxPackets, oldVtiRxPackets)
				deltaTxOctets = maxvalue(max, getDeltaMetrics(sumTxOctets, oldTxOctets))
				deltaVtiTxOctets = maxvalue(max, getDeltaMetrics(sumVtiTxOctets, oldVtiTxOctets))
				deltaRxOctets = maxvalue(max, getDeltaMetrics(sumRxOctets, oldRxOctets))
				deltaVtiRxOctets = maxvalue(max, getDeltaMetrics(sumVtiRxOctets, oldVtiRxOctets))
			}
		}

		var mac net.HardwareAddr = j.GetMacAddress()
		// if mac address field is empty, use "00:00:00:00:00:00" instead.
		if mac == nil {
			mac, _ = net.ParseMAC("00:00:00:00:00:00")
		}

		// Each traffic tuple will print two metrics log after discussing with Glenn and Yo.
		// This will reduce metrics log for each traffic tuple to 2, one is VTI, another is
		// non-VTI interface. Previously we could have one metrics log for each traffic volume
		// in traffic tuple.

		// The interface in the metrics log is no longer used but we will keep this field here
		// to maintain database backward compatibility. Also there is no need to update logstash and
		// elasticsearch configuration. Since orch use WAN1 and WAN2 to filter traffic in source
		// application and destination tab in monitor page, we will use WAN1 here to match the
		// orch elasticsearch search.
		inf = "WAN1"
		isVTI = false
		// Only print the log if has traffic volume on WAN or VTI interface
		if hasWAN {
			metrics.Printf("ReportTrafficInfo: Time: %v SN: %s "+
				"ClientIP: %v ClientPort: %v RemoteIP: %v RemotePort: %v Protocol: %v Interface: %v "+
				"Application: %v ClientHostname: %v ClientUsername: %v ClientOperatingSystem: %v Qos: %v "+
				"Status: %v MacAddress: %v isVTI: %v Magic: %v IcmpID: %v "+
				"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v "+
				"DeltaTxPackets: %v DeltaRxPackets: %v DeltaTxOctets: %v DeltaRxOctets: %v",
				t3339, sn, net.IP(j.GetClientIp()), j.GetClientPort(), net.IP(j.GetRemoteIp()), j.GetRemotePort(),
				j.GetProto(), inf, j.GetApplication(), j.GetClientHostname(), j.GetClientUsername(),
				j.GetClientOperatingSystem(), j.GetQos(), j.GetStatus(), mac.String(), isVTI, j.GetMagic(), j.GetIcmpid(),
				sumTxOctets, sumTxPackets, sumRxOctets, sumRxPackets,
				deltaTxPackets, deltaRxPackets, deltaTxOctets, deltaRxOctets)
		}

		inf = "WAN1"
		isVTI = true
		if hasVTI {
			metrics.Printf("ReportTrafficInfo: Time: %v SN: %s "+
				"ClientIP: %v ClientPort: %v RemoteIP: %v RemotePort: %v Protocol: %v Interface: %v "+
				"Application: %v ClientHostname: %v ClientUsername: %v ClientOperatingSystem: %v Qos: %v "+
				"Status: %v MacAddress: %v isVTI: %v Magic: %v IcmpID: %v "+
				"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v "+
				"DeltaTxPackets: %v DeltaRxPackets: %v DeltaTxOctets: %v DeltaRxOctets: %v",
				t3339, sn, net.IP(j.GetClientIp()), j.GetClientPort(), net.IP(j.GetRemoteIp()), j.GetRemotePort(),
				j.GetProto(), inf, j.GetApplication(), j.GetClientHostname(), j.GetClientUsername(),
				j.GetClientOperatingSystem(), j.GetQos(), j.GetStatus(), mac.String(), isVTI, j.GetMagic(), j.GetIcmpid(),
				sumVtiTxOctets, sumVtiTxPackets, sumVtiRxOctets, sumVtiRxPackets,
				deltaVtiTxPackets, deltaVtiRxPackets, deltaVtiTxOctets, deltaVtiRxOctets)
		}
		// the status is destroy, remove this connection from traffic info map
		if j.GetStatus() == pb.Status_DESTROY {
			glog.V(10).Infof("Destroy status received, deleted this connection %v from traffic info mapping", j)
			rti.deleteMapObject(key)
			continue
		}

		// add this traffic volume to hash map object
		k := copyNewTrafficTuple(j, sumTxPackets, sumRxPackets, sumTxOctets, sumRxOctets, sumVtiTxPackets, sumVtiRxPackets, sumVtiTxOctets, sumVtiRxOctets)
		rti.addMapObject(key, &k)

	}
	return nil
}

func maxvalue(max uint64, input uint64) uint64 {
	if input > max {
		return max
	}
	return input

}

// copyNewTrafficTuple copy the traffic tuple to new struct of traffic tuple which
// has only one traffic volume in the traffic tuple. This new struct aovid an o(n)
// search in the print the hash map of traffic tuple.
func copyNewTrafficTuple(m *pb.TrafficTuple, TxPackets uint64, RxPackets uint64, TxOctets uint64, RxOctets uint64, VtiTxPackets uint64, VtiRxPackets uint64, VtiTxOctets uint64, VtiRxOctets uint64) newTrafficTuple {
	var k newTrafficTuple
	k.ClientIP = m.GetClientIp()
	k.ClientPort = m.GetClientPort()
	k.Proto = m.GetProto()
	k.RemoteIP = m.GetRemoteIp()
	k.RemotePort = m.GetRemotePort()
	k.EstablishAt = m.GetEstablishAt()
	k.ReportAt = m.GetReportAt()
	k.Application = m.GetApplication()
	k.ClientHostname = m.GetClientHostname()
	k.ClientUsername = m.GetClientUsername()
	k.ClientOperatingSystem = m.GetClientOperatingSystem()
	k.Qos = m.GetQos()
	k.Status = m.GetStatus()
	k.TxPackets = TxPackets
	k.RxPackets = RxPackets
	k.TxOctets = TxOctets
	k.RxOctets = RxOctets
	k.VtiTxPackets = VtiTxPackets
	k.VtiRxPackets = VtiRxPackets
	k.VtiTxOctets = VtiTxOctets
	k.VtiRxOctets = VtiRxOctets
	k.MacAddress = m.GetMacAddress()
	return k
}

func logJSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	glog.V(10).Infoln(string(b))
	return nil
}

func convertString(m interface{}) (string, error) {
	out, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// getTime read google protobuf timestamp and return with string
// timestamp output with RFC3339 format.
func getTime(t *google_protobuf.Timestamp) string {
	var n time.Time

	if t == nil {
		n = time.Now()
	} else {
		n, _ = ptypes.Timestamp(t)
	}
	return n.Format(time.RFC3339)
}

// getTime read google protobuf timestamp and return with string
// timestamp output with RFC3339 format.
func getTimeNano(t *google_protobuf.Timestamp) string {
	var n time.Time

	if t == nil {
		n = time.Now()
	} else {
		n, _ = ptypes.Timestamp(t)
	}
	return n.Format(time.RFC3339Nano)
}

// getTime read google protobuf timestamp and return with string
// timestamp output with RFC3339 format.
func getSN(sn string) string {
	if sn == "" {
		sn = "none"
	}
	return sn
}

// printDeviceEvent accept an raw struct, convert it to string and
// print it to events log and stdout. For raw events which does NOT require
// additional processing, we could directly pass the message to logstash
// and let logstash process with it.
func printDeviceEvent(m interface{}, s string) error {
	str, err := convertString(m)
	if err != nil {
		return err
	}
	glog.V(10).Infoln(str)
	events.Println(s+":", str)
	return nil
}

// printDeviceEvent accept an raw struct, convert it to string and
// print it to events log and stdout. For raw metrics which does NOT require
// additional processing, we could directly pass the message to logstash and
// let logstash process with it.
func printDeviceMetric(m interface{}, s string) error {
	str, err := convertString(m)
	if err != nil {
		return err
	}
	glog.V(10).Infoln(str)
	metrics.Println(s+":", str)
	return nil
}

func printPortStateChange(m *pbevent.Ports) error {
	/*
		if err := logJSON(m); err != nil {
			return err
		}

		var n newPorts

		n.DeviceSn = m.GetDeviceSn()
		n.Timestamp = getTime(m.GetTimestamp())
		//siteadmin := apiclient.GetSNSiteAdmin(n.DeviceSn)

		// check if siteadmin is nil
		if siteadmin == nil {
			glog.Infof("No site has been found for SN %v", n.DeviceSn)
			return errors.New("No site found")
		}

		// check if site owner reference is nil
		if siteadmin.Site.GetOwnerReferences() == nil {
			glog.Infof("This site %v is a orphan site", n.DeviceSn)
			return errors.New("This is an orphan site")
		}

		n.Corp = siteadmin.Site.GetOwnerReferences()[0].Name
		n.Name = siteadmin.Site.Spec.Name
		n.SiteID = siteadmin.Site.Spec.SiteId

		// check if corp is nil
		if n.Corp == "" {
			glog.Infof("No corp has been found for SN %v", n.DeviceSn)
			return errors.New("No corp found")
		}

		// check if name is nil
		if n.Name == "" {
			glog.Infof("No site name has been found for SN %v", n.DeviceSn)
			return errors.New("No site name found")
		}

		// check if siteID is 0
		if n.SiteID == 0 {
			glog.Infof("No site ID has been found for SN %v", n.DeviceSn)
			return errors.New("No site ID found")
		}

		for _, v := range m.GetPort() {
			n.PortName = v.GetPortName()
			n.State = v.GetState()
			n.QueryKey = n.DeviceSn + "-" + n.PortName

			out, err := json.Marshal(n)
			if err != nil {
				events.Printf("Error: %v, Rawmsg: %v", err, m)
				return err
			}

			events.Println("PortStateChange: ", string(out))
		}
	*/
	return nil
}

func printSystemLoad(m *pb.SystemLoad) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var n newSystemLoad

	n.DeviceSn = m.GetDeviceSn()
	n.Timestamp = getTime(m.GetTimestamp())
	n.Memory = m.Memory
	n.Network = m.Network
	n.CPU.Utilization = make(map[string]uint32)
	n.CPU.Utilization1Min = make(map[string]uint32)
	n.CPU.Utilization5Min = make(map[string]uint32)

	// get the total core of cpu
	n.CPU.TotalCore = len(m.GetCpu().GetCore())

	// change cpu utilization to simplified version per orch request.
	for _, v := range m.GetCpu().GetCore() {
		key := "CPU" + fmt.Sprint(v.GetIndex())
		n.CPU.Utilization[key] = v.GetMetrics().GetUtilization()
		n.CPU.Utilization1Min[key] = v.GetMetrics().GetUtilization1Min()
		n.CPU.Utilization5Min[key] = v.GetMetrics().GetUtilization5Min()
	}

	out, err := json.Marshal(n)
	if err != nil {
		metrics.Printf("Error: %v, Rawmsg: %v", err, m)
		return err
	}

	metrics.Println("ReportSystemLoad: ", string(out))

	return nil
}

func printReportARP(m *pbevent.ARPTable) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var n newARPEntry

	n.DeviceSn = m.GetDeviceSn()
	n.Timestamp = getTime(m.GetTimestamp())

	for _, k := range m.GetEntry() {
		n.IPAddr = net.IP(k.GetIpAddr()).String()
		n.MacAddress = net.HardwareAddr(k.GetMacAddress()).String()
		n.Device = k.GetDevice()

		out, err := json.Marshal(n)
		if err != nil {
			metrics.Printf("Error: %v, Rawmsg: %v", err, m)
			return err
		}

		events.Println("ReportARPTable: ", string(out))
	}

	return nil
}

func printClientInfo(m *pbevent.ClientInfo) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var n newClientInfo

	n.DeviceSn = m.GetDeviceSn()
	n.Timestamp = getTime(m.GetTimestamp())
	n.Hostname = m.GetHostname()
	n.Ipv4Addr = m.GetIpv4Addr()
	n.Ipv6Addr = m.GetIpv6Addr()
	n.LoginName = m.GetLoginName()
	n.MacAddr = net.HardwareAddr(m.GetMacAddr()).String()

	// For orchestrastor better handling with data in ES, OsInfo has been merge into
	// ClientInfo structure.
	os := m.GetOs()
	n.DhcpPacket = os.GetDhcpPacket()
	n.ProductName = os.GetProductName()
	n.ProductVersion = os.GetProductVersion()
	n.ExtraOsInfo = os.GetExtraOsInfo()

	out, err := json.Marshal(n)
	if err != nil {
		metrics.Printf("Error: %v, Rawmsg: %v", err, m)
		return err
	}

	events.Println("ClientInfoUpdate: ", string(out))
	return nil
}

func printReportLiveInfo(m *pb.LiveReport) error {
	if err := logJSON(m); err != nil {
		return err
	}

	// For firmware does not upgrade to latest, add current timestamp
	t := getTimeNano(m.GetTimestamp())

	// For firmware does not include SN, insert an none SN
	sn := getSN(m.GetDeviceSn())

	for _, j := range m.Interfaces {
		isVTI := false
		// Convert interface vti00012WAN1-00011WAN1 into WAN1 as discussed with Zyxel
		inf := j.GetName()
		// if interface is vti interface, mark with isVTI for transport page VPN traffic
		if strings.Contains(inf, "vti") {
			isVTI = true
			re := regexp.MustCompile(`^vti\d{5}(WAN\d+(?:G\d+|I\d+)?|Cel\d+)-\d{5}(?:WAN\d+(?:G\d+|I\d+)?|Cel\d+)`)
			// Inf only convert to WAN or Cellular interface. Valid interface examples are:
			// vti00001WAN1-00002WAN2 will convert back to WAN1
			// vti00001Cel1-00002WAN1 will convert back to Cellular1
			// vti00001WAN1G1-00002WAN2G2 will convert back to WAN1-GRE1
			// vti00001WAN1I1-00002WAN2I2 will convert back to WAN1-IPSEC1
			subinf := re.FindStringSubmatch(inf)
			if len(subinf) > 0 {
				if matched, _ := regexp.MatchString(`^WAN\d+$`, subinf[1]); matched {
					inf = subinf[1]
				} else if matched, _ := regexp.MatchString(`^Cel\d+`, subinf[1]); matched {
					inf = strings.Replace(subinf[1], "Cel", "Cellular", -1)
				} else if matched, _ := regexp.MatchString(`^WAN\d+(?:G\d+|I\d+)`, subinf[1]); matched {
					if strings.Contains(subinf[1], "I") {
						inf = strings.Replace(subinf[1], "I", "-IPSEC", -1)
					} else if strings.Contains(subinf[1], "G") {
						inf = strings.Replace(subinf[1], "G", "-GRE", -1)
					}
				}
			}
		}

		metrics.Printf("ReportLiveInfo: Time: %v SN: %v Interface: %v "+
			"Jitter: %v Loss: %v Protocol: ICMP "+
			"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v isVTI: %v ",
			t, sn, inf,
			j.Metrics.GetJitterMilliseconds(), j.Metrics.GetPacketLost(),
			j.IcmpMetrics.GetOutOctets(), j.IcmpMetrics.GetOutPackets(), j.IcmpMetrics.GetInOctets(), j.IcmpMetrics.GetInPackets(), isVTI)
		metrics.Printf("ReportLiveInfo: Time: %v SN: %v Interface: %v "+
			"Jitter: %v Loss: %v Protocol: TCP "+
			"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v isVTI: %v ",
			t, sn, inf,
			j.Metrics.GetJitterMilliseconds(), j.Metrics.GetPacketLost(),
			j.TcpMetrics.GetOutOctets(), j.TcpMetrics.GetOutPackets(), j.TcpMetrics.GetInOctets(), j.TcpMetrics.GetInPackets(), isVTI)
		metrics.Printf("ReportLiveInfo: Time: %v SN: %v Interface: %v "+
			"Jitter: %v Loss: %v Protocol: UDP "+
			"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v isVTI: %v ",
			t, sn, inf,
			j.Metrics.GetJitterMilliseconds(), j.Metrics.GetPacketLost(),
			j.UdpMetrics.GetOutOctets(), j.UdpMetrics.GetOutPackets(), j.UdpMetrics.GetInOctets(), j.UdpMetrics.GetInPackets(), isVTI)
		metrics.Printf("ReportLiveInfo: Time: %v SN: %v Interface: %v "+
			"Jitter: %v Loss: %v Protocol: Other "+
			"TxOctets: %v TxPackets: %v RxOctets: %v RxPackets: %v isVTI: %v ",
			t, sn, inf,
			j.Metrics.GetJitterMilliseconds(), j.Metrics.GetPacketLost(),
			j.OtherMetrics.GetOutOctets(), j.OtherMetrics.GetOutPackets(), j.OtherMetrics.GetInOctets(), j.OtherMetrics.GetInPackets(), isVTI)
	}
	return nil
}

func printReportLinkQuality(m *pb.LinkQuality) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var n newPerLinkQuality

	n.DeviceSn = m.GetDeviceSn()
	n.Timestamp = getTime(m.GetTimestamp())

	// loop through perlinkquality in peers in the linkquality for orchestrator to process
	for _, i := range m.GetPeer() {
		n.PeerName = i.GetPeerName()
		for _, j := range i.GetPerLinkQuality() {
			n.LinkName = j.GetLinkName()
			n.QualityParameters = j.GetQualityParameters()
			n.Failure = j.GetFailure()

			out, err := json.Marshal(n)
			if err != nil {
				metrics.Printf("Error: %v, Rawmsg: %v", err, m)
				return err
			}

			metrics.Println("ReportLinkQuality: ", string(out))
		}
	}

	return nil
}

func printReportVPNTrafficInfo(m *pb.VPNTrafficInfo) error {
	if err := logJSON(m); err != nil {
		return err
	}

	// For firmware does not upgrade to latest, add current timestamp
	t := getTime(m.GetTimestamp())

	// For firmware does not include SN, insert an none SN
	sn := getSN(m.GetDeviceSn())

	for _, i := range m.GetConnInfo() {
		metrics.Printf("ReportVPNTrafficInfo: Time: %v SN: %s "+
			"ConnName: %v RxPackets: %v RxOctets: %v TxPackets: %v TxOctets: %v",
			t, sn, i.GetConnName(), i.GetIngress().GetPackets(), i.GetIngress().GetOctets(),
			i.GetEgress().GetPackets(), i.GetEgress().GetOctets())
	}
	return nil
}

func printVPNEvent(m *pbevent.VPNEvent) error {
	/*
		if err := logJSON(m); err != nil {
			return err
		}
		// For firmware does not upgrade to latest, add current timestamp
		t := getTime(m.GetTimestamp())

		// For firmware does not include SN, insert an none SN
		sn := getSN(m.GetDeviceSn())

		conn := m.GetConnectionInfo()
		ike := conn.GetIke()
		ipsec := conn.GetIpsec()
		peer := m.GetPeerInfo()

		if ike != nil {
			if ike.GetAlg() == "" {
				ike.Alg = "none"
			}
		}

		if conn != nil {
			if conn.GetTunnelName() == "" {
				glog.V(10).Infof("No tunnel name has been found for SN %v", sn)
				return errors.New("No tunnel name found")
			}
		}

		// if local or remote address is nil, use 0.0.0.0. So logstash could correctly send
		// message to elasticsearch
		localAddr := net.IP(conn.GetLocalAddr()).String()
		remoteAddr := net.IP(conn.GetRemoteAddr()).String()
		peerAddr := peer.GetPeerAddr()
		if localAddr == "<nil>" {
			localAddr = "0.0.0.0"
		}
		if remoteAddr == "<nil>" {
			remoteAddr = "0.0.0.0"
		}
		if peerAddr == "" {
			peerAddr = "0.0.0.0"
		}

		//siteadmin := apiclient.GetSNSiteAdmin(sn)

		//CUB-1428: Get peer site information for peer site name in alert email
		var id, peerSiteName string
		peerSiteName = "None"

		inf := conn.GetInterfaceName()
		re := regexp.MustCompile(`^vti\d{5}(?:WAN\d+(?:G\d+|I\d+)?|Cel\d+)-0*(\d{1,5})WAN\d+(?:G\d+|I\d+)?|Cel\d+`)
		found := re.FindStringSubmatch(inf)

		// peer site id match from VTI interface
		if len(found) == 2 {
			id = found[1]
			peerSiteIDuint64, err := strconv.ParseUint(id, 10, 64)
			// check if conversion has been failed
			if err != nil {
				glog.Errorf("Uint32 conversion failed for string in VPNEvent: %v", id)
			} else {
				peerSiteID := uint32(peerSiteIDuint64)
				//peerSiteadmin := apiclient.GetSiteIDSiteAdmin(peerSiteID)

				// check if peerSiteadmin is nil
				if peerSiteadmin == nil {
					glog.Infof("No peer site has been found for site id %v", peerSiteID)
					peerSiteName = "NoPeerSiteName"
				} else {
					// check if site name is nil
					if peerSiteadmin.Site.Spec.Name != "" {
						peerSiteName = peerSiteadmin.Site.Spec.Name
					}
				}
			}
		}

		// check if siteadmin is nil
		if siteadmin == nil {
			glog.Infof("No site has been found for SN %v", sn)
			return errors.New("No site found")
		}

		// check if site owner reference is nil
		if siteadmin.Site.GetOwnerReferences() == nil {
			glog.Infof("This site %v is a orphan site", sn)
			return errors.New("This is an orphan site")
		}

		corp := siteadmin.Site.GetOwnerReferences()[0].Name
		name := siteadmin.Site.Spec.Name
		siteID := siteadmin.Site.Spec.SiteId

		// check if corp is nil
		if corp == "" {
			glog.Infof("No corp has been found for SN %v", sn)
			return errors.New("No corp found")
		}

		// check if name is nil
		if name == "" {
			glog.Infof("No site name has been found for SN %v", sn)
			return errors.New("No site name found")
		}

		// check if siteID is 0
		if siteID == 0 {
			glog.Infof("No site ID has been found for SN %v", sn)
			return errors.New("No site ID found")
		}

		events.Printf("VPNEvent: Time: %v SN: %s "+
			"Corp: %v Name: %v SiteID: %v EventType: %v TunnelName: %v Interface: %v VPNType: %v "+
			"LocalAddr: %v LocalPort: %v RemoteAddr: %v RemotePort: %v LocalPolicy: %v RemotePolicy: %v "+
			"IKE: spi_i: %v spi_r: %v ALG: %v CipherKey: %v IPSEC: spi_i: %v spi_o: %v PeerHandle: %v "+
			"PeerAddr: %v PeerIsServer: %v Uptime: %v TimeOut: %v Msg: %v",
			t, sn, corp, name, siteID, m.GetEventType(), conn.GetTunnelName(), conn.GetInterfaceName(), conn.GetVpnType(),
			localAddr, conn.GetLocalPort(), remoteAddr, conn.GetRemotePort(),
			conn.GetLocalPolicy().GetPolicyMode(), conn.GetRemotePolicy().GetPolicyMode(),
			ike.GetSpiI(), ike.GetSpiR(), ike.GetAlg(), ike.GetCipherKey(),
			ipsec.GetSpiI(), ipsec.GetSpiO(), ipsec.GetPeerHandle(),
			peerAddr, peer.GetPeerIsServer(), m.GetUptime(), m.GetTimeout(), m.GetMsg())

		events.Printf("VPNAlert: Time: %v SN: %s "+
			"Corp: %v Name: %v SiteID: %v EventType: %v TunnelName: %v Interface: %v VPNType: %v "+
			"LocalAddr: %v LocalPort: %v RemoteAddr: %v RemotePort: %v LocalPolicy: %v RemotePolicy: %v "+
			"IKE: spi_i: %v spi_r: %v ALG: %v CipherKey: %v IPSEC: spi_i: %v spi_o: %v PeerHandle: %v "+
			"PeerSiteName: %v PeerAddr: %v PeerIsServer: %v Uptime: %v TimeOut: %v Msg: %v",
			t, sn, corp, name, siteID, m.GetEventType(), conn.GetTunnelName(), conn.GetInterfaceName(), conn.GetVpnType(),
			localAddr, conn.GetLocalPort(), remoteAddr, conn.GetRemotePort(),
			conn.GetLocalPolicy().GetPolicyMode(), conn.GetRemotePolicy().GetPolicyMode(),
			ike.GetSpiI(), ike.GetSpiR(), ike.GetAlg(), ike.GetCipherKey(),
			ipsec.GetSpiI(), ipsec.GetSpiO(), ipsec.GetPeerHandle(),
			peerSiteName, peerAddr, peer.GetPeerIsServer(), m.GetUptime(), m.GetTimeout(), m.GetMsg())
	*/
	return nil
}

func printReportDeviceLog(m pbevent.CubsEventReport_ReportDeviceLogServer) error {
	glog.Info("Start syslog streaming.")
	for {
		in, err := m.Recv()
		if err == io.EOF {
			glog.Info("Receive EOF and stop syslog streaming.")
			return m.SendAndClose(emptyRsp)
		}
		if err != nil {
			return err
		}

		printDeviceLog(in)
	}
}

func printDeviceLog(m *pbevent.Devicelog) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var n newDevicelog

	n.DeviceSn = m.GetDeviceSn()
	n.Timestamp = m.GetTimestamp()
	n.Severity = m.GetSeverity()
	n.Facility = m.GetFacility()
	n.Category = m.GetCategory()
	n.Srcip = net.IP(m.GetSrcip()).String()
	n.Dstip = net.IP(m.GetDstip()).String()
	n.Ipproto = m.GetIpproto()
	n.Sport = m.GetSport()
	n.Dport = m.GetDport()
	n.Logmessage = m.GetLogmessage()
	n.Note = m.GetNote()
	n.Username = m.GetUsername()
	n.Srciface = m.GetSrciface()
	n.Dstiface = m.GetDstiface()
	n.ProtoName = m.GetProtoName()
	n.Devmac = m.GetDevmac()
	n.Count = m.GetCount()
	n.TrafficLog = m.GetTrafficLog()
	n.IdpLog = m.GetIdpLog()
	n.FirewallLog = m.GetFirewallLog()
	n.GeoSrc = m.GetGeoSrc()
	n.GeoDst = m.GetGeoDst()

	out, err := json.Marshal(n)
	if err != nil {
		events.Printf("Error: %v, Rawmsg: %v", err, m)
		return err
	}

	events.Println("Syslog: ", string(out))
	return nil
}

func printReportDNSAnswer(m pbevent.CubsEventReport_ReportDNSAnswerServer) error {
	glog.Info("Start ReportDnsAnswer streaming.")
	for {
		in, err := m.Recv()
		if err == io.EOF {
			glog.Info("Receive EOF and stop ReportDnsAnswer streaming.")
			return m.SendAndClose(emptyRsp)
		}
		if err != nil {
			return err
		}

		printDNS(in)
	}
}

func printDNS(m *pbevent.DnsAnswer) error {
	if err := logJSON(m); err != nil {
		return err
	}

	var n newDNSAnswer

	n.DeviceSn = m.GetDeviceSn()
	n.Timestamp = getTime(m.GetTimestamp())

	nSlice := make([]*newDNSAnswerEntry, 0, len(m.Entry))

	for _, p := range m.Entry {
		q := &newDNSAnswerEntry{}
		q.Addr = net.IP(p.GetAddr()).String()
		q.Cls = p.GetCls()
		q.Fqdn = p.GetFqdn()
		q.Name = p.GetName()
		q.TTL = p.GetTtl()
		q.Type = p.GetType()
		// only add the entry if the address is not nil
		if q.Addr != "<nil>" {
			nSlice = append(nSlice, q)
		}
	}

	// if nSlice is empty slice, do not write to event log
	if len(nSlice) == 0 {
		return nil
	}
	n.Entry = nSlice
	out, err := json.Marshal(n)
	if err != nil {
		events.Printf("Error: %v, Rawmsg: %v", err, m)
		return err
	}

	events.Println("ReportDnsAnswer: ", string(out))
	return nil
}

// func composeNetTuple compose flow tuples and return it with string
func composeNetTuple(m *pb.TrafficTuple) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v_%v", net.IP(m.GetClientIp()), m.GetClientPort(),
		net.IP(m.GetRemoteIp()), m.GetRemotePort(), m.GetProto(), m.GetMagic())
}

// func composeNetTuple compose flow tuples and return it with string
func composeNetTupleWOMagic(m *pb.TrafficTuple) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v", net.IP(m.GetClientIp()), m.GetClientPort(),
		net.IP(m.GetRemoteIp()), m.GetRemotePort(), m.GetProto())
}
