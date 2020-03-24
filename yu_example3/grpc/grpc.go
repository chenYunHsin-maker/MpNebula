package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	//"strconv"

	//"sync"
	"time"

	pbevent "yu_example3/events"

	//"google.golang.org/grpc"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	proxyproto "github.com/armon/go-proxyproto"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/timestamp"

	//"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	//"sdn.io/sdwan/cmd/cubs/monitorproxy/apiclient"
	//"sdn.io/sdwan/cmd/cubs/monitorproxy/elastic"
	//pbevent "yu_example3/events"
	pb "yu_example3/metrics"
)

var (
	grpcServer *grpc.Server
	Port       int
	metrics    *log.Logger
	events     *log.Logger
	Cert       string
	Key        string
	Ca         string
)

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle: 5 * time.Minute,
	Time:              2 * time.Hour,    // Ping the client if it is idle for 2 hours to ensure the connection is still active
	Timeout:           20 * time.Second, // Wait 20 seconds for the ping ack before assuming the connection is dead

}
var kaspClient = keepalive.ClientParameters{
	Time:    2 * time.Hour,    // Ping the client if it is idle for 2 hours to ensure the connection is still active
	Timeout: 20 * time.Second, // Wait 20 seconds for the ping ack before assuming the connection is dead
}

// StopServer stops the gRPC server
func StopServer() {
	grpcServer.Stop()
}

// GracefulStopServer gracefully stops the gRPC server
func GracefulStopServer() {
	grpcServer.GracefulStop()
}

// Define empty cubs metrics, events and turboproxy server structs
type cubsServer struct{}
type cubsEventReportServer struct{}
type turboproxyServer struct{}

// Define empty response for events and metrics
var empty = &pb.Empty{}
var emptyRsp = &pbevent.Empty{}

// Define new cubs metrics, event and turboproxy server func
func newCubsServer() *cubsServer {
	s := new(cubsServer)
	return s
}

func newCubsEventReportServer() *cubsEventReportServer {
	s := new(cubsEventReportServer)
	return s
}

func newTurboproxyServer() *turboproxyServer {
	s := new(turboproxyServer)
	return s
}

func (s *cubsServer) ReportVPNTrafficInfo(ctx context.Context, m *pb.VPNTrafficInfo) (*pb.Empty, error) {
	if err := printReportVPNTrafficInfo(m); err != nil {
		return empty, err
	}
	return empty, nil
}

// Func ReportTrafficInfo is used for monitoring history traffic based on SN, flow tuples and interfaces
func (s *cubsServer) ReportTrafficInfo(ctx context.Context, m *pb.AccumulatedTraffic) (*pb.Empty, error) {
	if err := printReportTrafficInfo(m); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *cubsServer) ReportFlowTrafficInfo(ctx context.Context, m *pb.FlowTrafficInfo) (*pb.Empty, error) {
	if err := printFlowTrafficInfo(m); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *cubsServer) ReportLinkQuality(ctx context.Context, m *pb.LinkQuality) (*pb.Empty, error) {
	if err := printReportLinkQuality(m); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *cubsServer) ReportLiveInfo(ctx context.Context, m *pb.LiveReport) (*pb.Empty, error) {
	if err := printReportLiveInfo(m); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *cubsServer) ReportSystemLoad(ctx context.Context, m *pb.SystemLoad) (*pb.Empty, error) {
	if err := printSystemLoad(m); err != nil {
		return empty, err
	}
	return empty, nil
}

// client
func (s *cubsEventReportServer) ClientInfoUpdate(ctx context.Context, m *pbevent.ClientInfo) (*pbevent.Empty, error) {
	if err := printClientInfo(m); err != nil {
		fmt.Println("cub server:", err)
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) ClientTimeout(ctx context.Context, m *pbevent.ClientInfo) (*pbevent.Empty, error) {
	if err := printClientInfo(m); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) UserLogin(ctx context.Context, m *pbevent.UserInfo) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "UserLogin"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) UserLogout(ctx context.Context, m *pbevent.UserInfo) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "UserLogout"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// interfaces
func (s *cubsEventReportServer) IPChanged(ctx context.Context, m *pbevent.Interfaces) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "IPChanged"); err != nil {
		return emptyRsp, err
	}
	//apiclient.IpchangedCallback(m)
	return emptyRsp, nil
}

func (s *cubsEventReportServer) LinkStateChange(ctx context.Context, m *pbevent.Interfaces) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "LinkStateChange"); err != nil {
		return emptyRsp, err
	}

	return emptyRsp, nil
}

func (s *cubsEventReportServer) InterfaceStatusChange(ctx context.Context, m *pbevent.Interfaces) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "InterfaceStatusChange"); err != nil {
		return emptyRsp, err
	}

	return emptyRsp, nil
}

func (s *cubsEventReportServer) LivemonStopped(ctx context.Context, m *pbevent.Interface) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "LivemonStopped"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) PortStateChange(ctx context.Context, m *pbevent.Ports) (*pbevent.Empty, error) {
	if err := printPortStateChange(m); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// system
func (s *cubsEventReportServer) FirmwareDownloadDone(ctx context.Context, m *pbevent.FirmwareDownloaded) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "FirmwareDownloadDone"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) FirmwareDownloadReport(ctx context.Context, m *pbevent.FirmwareDownloadProcess) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "FirmwareDownloadReport"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) FirmwareUpgradeStart(ctx context.Context, m *pbevent.FirmwareUpgradeStarted) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "FirmwareUpgradeStart"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) SystemAlert(ctx context.Context, m *pbevent.SystemAlertMessage) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "SystemAlert"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) PackageUpgradeStart(ctx context.Context, m *pbevent.PackageUpgradeStartMessage) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "PackageUpgradeStart"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) PackageUpgradeResult(ctx context.Context, m *pbevent.PackageUpgradeResultMessage) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "PackageUpgradeResult"); err != nil {
		return emptyRsp, err
	}
	//apiclient.PkgUpgradeCallback(m)
	return emptyRsp, nil
}

func (s *cubsEventReportServer) DiskLogUpload(ctx context.Context, m *pbevent.DiskLogUploadMessage) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "DiskLogUploadMessage"); err != nil {
		return emptyRsp, err
	}
	//apiclient.DiskLogUploadCallback(m)
	return emptyRsp, nil
}

// usb
func (s *cubsEventReportServer) USBDeviceDetected(ctx context.Context, m *pbevent.USBDeviceInfo) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "USBDeviceDetected"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) USBDeviceRemoved(ctx context.Context, m *pbevent.USBDeviceInfo) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "USBDeviceRemoved"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// vpn
func (s *cubsEventReportServer) VPNTestDone(ctx context.Context, m *pbevent.VPNTestResult) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "VPNTestDone"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) OnVPNEvent(ctx context.Context, m *pbevent.VPNEvent) (*pbevent.Empty, error) {
	if err := printVPNEvent(m); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// debug-tun
func (s *cubsEventReportServer) TunnelEstablished(ctx context.Context, m *pbevent.DebugTunnelInfo) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "TunnelEstablished"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) TunnelDisconnected(ctx context.Context, m *pbevent.DebugTunnelInfo) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "TunnelDisconnected"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// debug-runcmd

func (s *cubsEventReportServer) CommandExecuted(ctx context.Context, m *pbevent.DebugCommand) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "CommandExecuted"); err != nil {
		return emptyRsp, err
	}
	//apiclient.CommandExecutedCallback(m)
	return emptyRsp, nil
}

// ping done
func (s *cubsEventReportServer) PingDone(ctx context.Context, m *pbevent.PingReport) (*pbevent.Empty, error) {
	/*
		if err := elastic.PrintPingDoneEvent(m); err != nil {
			glog.Errorln("PrintPingDoneEvent error :", err)
		}*/

	if err := printDeviceEvent(m, "PingDone"); err != nil {
		glog.Errorln("printDeviceEvent error :", err)
		return emptyRsp, err
	}
	fmt.Println("report:", m)

	//start client
	if Port != 10000 {
		//start certificate
		/*
			certificate, err := tls.LoadX509KeyPair(Cert, Key)

			certPool := x509.NewCertPool()
			bs, err := ioutil.ReadFile(Ca)
			if err != nil {
				glog.Fatalf("failed to read ca cert: %s", err)
			}

			ok := certPool.AppendCertsFromPEM(bs)
			if !ok {
				glog.Fatal("failed to append certs")
			}

			creds := credentials.NewTLS(&tls.Config{
				//ServerName: "localhost",
				ServerName:   "0.0.0.0",
				Certificates: []tls.Certificate{certificate},
				RootCAs:      certPool,
			})

			dialOption := grpc.WithTransportCredentials(creds)
		*/
		//end certificate

		//conn, err := grpc.Dial("ncc-mp2-monitorproxy:10000", dialOption)
		conn, err := grpc.Dial("ncc-mp2-monitorproxy:10000", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("連線失敗：%v", err)
		}
		defer conn.Close()
		//	now := time.Now()
		d := pbevent.NewCubsEventReportClient(conn)
		//d := pbevent.NewPingServerClient(conn)
		fmt.Println("ping 10000!")
		r, err := d.PingDone(context.Background(), &pbevent.PingReport{
			DeviceSn: "S1234567890",
			TaskId:   "1234",
			Report:   "so321a",
			Timestamp: &timestamp.Timestamp{
				Seconds: 12334567867,
				Nanos:   444444,
			},
		})
		if err != nil {
			log.Fatalf("無法執行 Plus 函式：%v", err)
		}
		log.Printf("回傳結果：%s ", r)

	}

	//end client
	return emptyRsp, nil
}

// traceroute done
func (s *cubsEventReportServer) TracerouteDone(ctx context.Context, m *pbevent.TracerouteReport) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "TracerouteDone"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// dns query done
func (s *cubsEventReportServer) DNSQueryDone(ctx context.Context, m *pbevent.DNSReport) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "DNSQueryDone"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// pcap stop
func (s *cubsEventReportServer) PacketCaptureStopped(ctx context.Context, m *pbevent.PacketCaptureReport) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "PacketCaptureStopped"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) PacketCaptureUploaded(ctx context.Context, m *pbevent.PacketCaptureUploadedMessage) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "PacketCaptureUploaded"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// list-path-done
func (s *cubsEventReportServer) ListPathDone(ctx context.Context, m *pbevent.ListPathReport) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "ListPathDone"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

// syslog streaming
func (s *cubsEventReportServer) ReportDeviceLog(m pbevent.CubsEventReport_ReportDeviceLogServer) error {
	return printReportDeviceLog(m)
}

func (s *cubsEventReportServer) ReportDNSAnswer(m pbevent.CubsEventReport_ReportDNSAnswerServer) error {
	return printReportDNSAnswer(m)
}

func (s *cubsEventReportServer) ReportARPTable(ctx context.Context, m *pbevent.ARPTable) (*pbevent.Empty, error) {
	if err := printReportARP(m); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) ReportWanoptNetstat(ctx context.Context, m *pbevent.DebugWanoptNetstat) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "DebugWanoptNetstat"); err != nil {
		return emptyRsp, err
	}
	return emptyRsp, nil
}

func (s *cubsEventReportServer) OnCGNATEvent(ctx context.Context, m *pbevent.CGNATEvent) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "CGNATEvent"); err != nil {
		return emptyRsp, err
	}
	//apiclient.CGNATCallback(m)
	return emptyRsp, nil
}

func (s *cubsEventReportServer) ReportDeviceHAStatus(ctx context.Context, m *pbevent.DeviceHAStatus) (*pbevent.Empty, error) {
	if err := printDeviceEvent(m, "ReportDeviceHAStatus"); err != nil {
		return emptyRsp, err
	}
	//apiclient.DeviceHaReportCallback(m)
	return emptyRsp, nil
}

// ReportSCTP reports SCTP statistics of an Edge at a time
func (s *turboproxyServer) ReportSCTP(ctx context.Context, m *pb.SCTPStat) (*pb.Empty, error) {
	if err := printDeviceEvent(m, "ReportSCTP"); err != nil {
		return empty, err
	}
	return empty, nil
}

// StartServer starts the gRPC server
func StartServer(sCertFile, sCertKeyFile, caFile string, port int, metricLogFile, eventLogFile string) {
	//var mu sync.Mutex
	//var wg sync.WaitGroup
	//mu.Lock()
	Cert = sCertFile
	Key = sCertKeyFile
	Ca = caFile
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	Port = port
	fmt.Println("server listen:", "0.0.0.0", port)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}

	cer, err := tls.LoadX509KeyPair(sCertFile, sCertKeyFile)
	if err != nil {
		glog.Fatal(err)
	}
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		glog.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cer},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Create log backend for metrics and events
	metricLog, err := os.OpenFile(metricLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Fatalf("Failed to open metrics log file: %v", err)
	}
	defer metricLog.Close()

	eventLog, err := os.OpenFile(eventLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Fatalf("Failed to open event log file: %v", err)
	}
	defer eventLog.Close()

	// Initialize log and set format
	metrics = log.New(metricLog, "Metrics: ", log.Ldate|log.Ltime|log.Lshortfile)
	// event log format require more precision (microsecond) to generate alert correctly
	events = log.New(eventLog, "Events: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	// Set log rotations
	metrics.SetOutput(&lumberjack.Logger{
		Filename:   metricLogFile,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})

	events.SetOutput(&lumberjack.Logger{
		Filename:   eventLogFile,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})

	creds := credentials.NewTLS(tlsConfig)
	// opts := []grpc.ServerOption{grpc.KeepaliveParams(kasp), grpc.Creds(creds)}
	fmt.Println("creds:", creds)
	opts := []grpc.ServerOption{grpc.Creds(creds), grpc.KeepaliveParams(kasp)}
	grpcServer = grpc.NewServer(opts...)
	// Register metrics grpc server on MP
	pb.RegisterCubsServer(grpcServer, newCubsServer())

	// Register turboproxy grpc server on MP
	pb.RegisterTurboProxyServer(grpcServer, newTurboproxyServer())

	// Register events grpc server on MP
	pbevent.RegisterCubsEventReportServer(grpcServer, newCubsEventReportServer())
	proxyLis := &proxyproto.Listener{Listener: lis}
	deleteOldTrafficInfo()
	//fmt.Println("proxylist:", proxyLis)
	grpcServer.Serve(proxyLis)
	//	mu.Unlock()
	//	wg.Done()
}
func StartServer2(sCertFile, sCertKeyFile, caFile string, port int, metricLogFile, eventLogFile string) {
	//var mu sync.Mutex
	//var wg sync.WaitGroup
	//mu.Lock()
	Cert = sCertFile
	Key = sCertKeyFile
	Ca = caFile
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	Port = port
	fmt.Println("server listen:", "0.0.0.0", port)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}

	cer, err := tls.LoadX509KeyPair(sCertFile, sCertKeyFile)
	if err != nil {
		glog.Fatal(err)
	}
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		glog.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cer},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Create log backend for metrics and events
	metricLog, err := os.OpenFile(metricLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Fatalf("Failed to open metrics log file: %v", err)
	}
	defer metricLog.Close()

	eventLog, err := os.OpenFile(eventLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Fatalf("Failed to open event log file: %v", err)
	}
	defer eventLog.Close()

	// Initialize log and set format
	metrics = log.New(metricLog, "Metrics: ", log.Ldate|log.Ltime|log.Lshortfile)
	// event log format require more precision (microsecond) to generate alert correctly
	events = log.New(eventLog, "Events: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	// Set log rotations
	metrics.SetOutput(&lumberjack.Logger{
		Filename:   metricLogFile,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})

	events.SetOutput(&lumberjack.Logger{
		Filename:   eventLogFile,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})

	creds := credentials.NewTLS(tlsConfig)
	// opts := []grpc.ServerOption{grpc.KeepaliveParams(kasp), grpc.Creds(creds)}
	fmt.Println("creds:", creds)
	opts := []grpc.ServerOption{grpc.KeepaliveParams(kasp)}
	grpcServer = grpc.NewServer(opts...)
	// Register metrics grpc server on MP
	pb.RegisterCubsServer(grpcServer, newCubsServer())

	// Register turboproxy grpc server on MP
	pb.RegisterTurboProxyServer(grpcServer, newTurboproxyServer())

	// Register events grpc server on MP
	pbevent.RegisterCubsEventReportServer(grpcServer, newCubsEventReportServer())
	proxyLis := &proxyproto.Listener{Listener: lis}
	deleteOldTrafficInfo()
	//fmt.Println("proxylist:", proxyLis)
	grpcServer.Serve(proxyLis)
	//	mu.Unlock()
	//	wg.Done()
}
