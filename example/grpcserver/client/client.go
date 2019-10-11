/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/emiraganov/pubsub/example/grpcserver/mymodels"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := c.GetEvents(ctx, &pb.Subscribe{
		App:    name,
		Agents: []string{"_ALL_"},
		Queues: []string{"_ALL_"},
	})
	if err != nil {
		log.Fatalf("could not get events: %v", err)
	}

	for {
		e := &pb.Event{}
		if err := stream.RecvMsg(e); err != nil {
			log.Fatalf("Receving msg: %v", err)
		}

		log.Println(e)

		var data proto.Message = &pb.TestEvent{}

		// var data proto.Message = pb.TestEvent{}
		if err := ptypes.UnmarshalAny(e.Data, data); err != nil {
			log.Println(err)
		}

		log.Println(data)
	}
}
