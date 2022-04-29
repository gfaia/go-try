package main

import (
	"context"
	"flag"
	"io"
	"jaeger/pkg/trace"
	"log"
	"net"

	testpb "jaeger/api/grpc_testing"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"google.golang.org/grpc"
)

var (
	reportAddr  string
	serviceName string
	grpcAddr    string
)

func init() {
	flag.StringVar(&reportAddr, "report_addr", "127.0.0.1:6831", "the address of jaeger agent.")
	flag.StringVar(&serviceName, "service_name", "server", "the name of service.")
	flag.StringVar(&grpcAddr, "grpc_addr", ":10000", "the address of grpc server.")
	flag.Parse()
}

const (
	streamLength = 5
)

type testServer struct {
	testpb.UnimplementedTestServiceServer
}

func (s *testServer) UnaryCall(ctx context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error) {
	return &testpb.SimpleResponse{Payload: in.Payload}, nil
}

func (s *testServer) StreamingOutputCall(in *testpb.SimpleRequest, stream testpb.TestService_StreamingOutputCallServer) error {
	for i := 0; i < streamLength; i++ {
		if err := stream.Send(&testpb.SimpleResponse{Payload: in.Payload}); err != nil {
			return err
		}
	}
	return nil
}

func (s *testServer) StreamingInputCall(stream testpb.TestService_StreamingInputCallServer) error {
	sum := ""
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		sum += in.Payload
	}
	return stream.SendAndClose(&testpb.SimpleResponse{Payload: sum})
}

func (s *testServer) StreamingBidirectionalCall(stream testpb.TestService_StreamingBidirectionalCallServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err = stream.Send(&testpb.SimpleResponse{Payload: in.Payload}); err != nil {
			return err
		}
	}
}

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	defer closer.Close()
	if err != nil {
		log.Printf("NewJaegerTracer err: %v", err.Error())
	}
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)),
		grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)))
	testpb.RegisterTestServiceServer(s, &testServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return
}
