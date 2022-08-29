package main

import (
	"context"
	"flag"
	"fmt"
	"jaeger/pkg/trace"
	"log"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	reportAddr  string
	serviceName string
	serviceAddr string
)

func init() {
	flag.StringVar(&reportAddr, "report_addr", "127.0.0.1:6831", "the address of jaeger agent.")
	flag.StringVar(&serviceName, "service_name", "server", "the name of service.")
	flag.StringVar(&serviceAddr, "service_addr", ":8181", "")
	flag.Parse()
}

func traceHandler(tracer opentracing.Tracer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract span context from http headers.
		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Println(err)
		}
		span := tracer.StartSpan("start", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		ctx = opentracing.ContextWithSpan(ctx, span)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "test")
		return
	}
}

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	http.HandleFunc("/", traceHandler(tracer))

	log.Fatalln(http.ListenAndServe(serviceAddr, nil))
	return
}
