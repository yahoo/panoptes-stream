package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry/mock"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/simulator/juniper"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ln, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	gServer := grpc.NewServer()
	juniperGnmiUpdate := juniper.New()
	mockServer := &mock.GNMIServer{Resp: juniperGnmiUpdate}
	gnmi.RegisterGNMIServer(gServer, mockServer)

	gServer.Serve(ln)
}
