package main

import (
	"context"
	"log"

	pbevent "yu_example3/echospec/events"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("連線失敗：%v", err)
	}
	defer conn.Close()
	//	now := time.Now()

	d := pbevent.NewPingServerClient(conn)
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

	/*c := pbevent.NewEchoClient(conn)
	r, err = c.Echo(context.Background(), &pbevent.EchoRequest{Msg: "HI HI HI HI"})
	if err != nil {
		log.Fatalf("無法執行 Plus 函式：%v", err)
	}
	log.Printf("回傳結果：%s , 時間:%d", r.Msg, r.Unixtime)*/

}
