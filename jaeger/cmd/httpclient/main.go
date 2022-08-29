package main

import (
	"context"
	"flag"
	"jaeger/pkg/trace"
	"log"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

var (
	reportAddr  string
	serviceName string
	serviceAddr string
)

func init() {
	flag.StringVar(&reportAddr, "report_addr", "127.0.0.1:6831", "the address of jaeger agent.")
	flag.StringVar(&serviceName, "service_name", "client", "the name of service.")
	flag.StringVar(&serviceAddr, "service_addr", ":8181", "")
	flag.Parse()
}

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a root span.
	span := tracer.StartSpan("start")
	defer span.Finish()

	// Set span into context.
	ctx = opentracing.ContextWithSpan(ctx, span)
	client := &http.Client{Transport: &nethttp.Transport{}}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8181", nil)
	if err != nil {
		log.Println(err)
		return
	}
	req = req.WithContext(ctx) // extend existing trace, if any

	req, ht := nethttp.TraceRequest(tracer, req)
	defer ht.Finish()

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	return
}
