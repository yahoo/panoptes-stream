//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry/mock"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/simulator/juniper"
)

func main() {
	var interval int

	rand.Seed(time.Now().UnixNano())

	flag.IntVar(&interval, "interval", 5, "update interval in seconds")
	flag.Parse()

	ln, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	gServer := grpc.NewServer()
	juniperGnmiUpdate := juniper.New(context.Background(), interval)
	mockServer := &mock.GNMIServer{Resp: juniperGnmiUpdate}
	gnmi.RegisterGNMIServer(gServer, mockServer)

	gServer.Serve(ln)
}
