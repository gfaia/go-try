package main

import (
	"context"
	"flag"
	"io"
	"jaeger/pkg/trace"
	"log"
	"time"

	testpb "jaeger/api/grpc_testing"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"google.golang.org/grpc"
)

var (
	reportAddr  string
	serviceName string
	grpcAddr    string
)

const (
	streamLength = 5
)

func init() {
	flag.StringVar(&reportAddr, "report_addr", "127.0.0.1:6831", "the address of jaeger agent.")
	flag.StringVar(&serviceName, "service_name", "client", "the name of service.")
	flag.StringVar(&grpcAddr, "grpc_addr", ":10000", "the address of grpc server.")
	flag.Parse()
}

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	if err != nil {
		log.Printf("NewJaegerTracer err: %s", err.Error())
	}
	defer closer.Close()

	conn, err := grpc.Dial(
		grpcAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
		grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := testpb.NewTestServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := c.UnaryCall(ctx, &testpb.SimpleRequest{Payload: `test`})
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("payload: %s", r.Payload)

	s, err := c.StreamingOutputCall(ctx, &testpb.SimpleRequest{Payload: `test`})
	if err != nil {
		log.Fatalf(err.Error())
	}
	for {
		resp, err := s.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed StreamingOutputCall: %v", err)
		}
		log.Printf("payload: %s", resp.Payload)
	}

	s2, err := c.StreamingInputCall(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}
	for i := 0; i < streamLength; i++ {
		if err = s2.Send(&testpb.SimpleRequest{Payload: `test`}); err != nil {
			log.Fatalf("Failed StreamingInputCall: %v", err)
		}
	}
	resp2, err := s2.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed StreamingInputCall: %v", err)
	}
	log.Printf("payload: %s", resp2.Payload)

	s3, err := c.StreamingBidirectionalCall(ctx)
	go func() {
		for i := 0; i < streamLength; i++ {
			if err := s3.Send(&testpb.SimpleRequest{Payload: `test`}); err != nil {
				log.Fatalf("Failed StreamingInputCall: %v", err)
			}
		}
		s3.CloseSend()
	}()
	for {
		resp3, err := s3.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed StreamingOutputCall: %v", err)
		}
		log.Printf("payload: %s", resp3.Payload)
	}
	return
}
