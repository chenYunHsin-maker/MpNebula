package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pbevent "yu_example3/events"

	"github.com/golang/glog"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	ws, _ := os.Getwd()
	//ws = os.Getenv("GOPATH") + "/src/sdn.io/sdwan"

	fmt.Println("hi")

	flag.StringVar(&cert, "cert", ws+"/certs/mycerts/old/server.pem", "The TLS cert file")
	flag.StringVar(&key, "key", ws+"/certs/mycerts/old/server-key.pem", "The TLS key file")
	flag.StringVar(&ca, "ca", ws+"/certs/mycerts/old/ca.pem", "The CA cert file")
	flag.IntVar(&port, "port", 9999, "The server port")
	flag.StringVar(&metricLog, "metric", "../grpc/testdata/metrics.log", "Metric log file")
	flag.StringVar(&eventLog, "event", "../grpc/testdata/events.log", "Event log file")
	flag.StringVar(&alertLog, "alert", "../grpc/testdata/alerts.log", "Alert log file")

	flag.Set("logtostderr", "true")
}
func main() {
	//start cert
	certificate, err := tls.LoadX509KeyPair(cert, key)

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(ca)
	if err != nil {
		glog.Fatalf("failed to read ca cert: %s", err)
	}

	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		glog.Fatal("failed to append certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		//ServerName:   "localhost",
		ServerName:   "127.0.0.1",
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	dialOption := grpc.WithTransportCredentials(creds)
	//end cert
	//conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:9999", dialOption)
	if err != nil {
		log.Fatalf("連線失敗：%v", err)
	}
	defer conn.Close()
	//	now := time.Now()
	d := pbevent.NewCubsEventReportClient(conn)
	//d := pbevent.NewPingServerClient(conn)
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
		log.Fatalf("client無法執行 Plus 函式：%v", err)
	}
	log.Printf("回傳結果：%s ", r)

}
