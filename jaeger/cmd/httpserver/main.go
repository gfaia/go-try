package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"jaeger/pkg/trace"

	nethttp "github.com/opentracing-contrib/go-stdlib/nethttp"
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

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	mux := http.NewServeMux()
	mux.Handle("/", nethttp.MiddlewareFunc(tracer, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test")
	}))
	log.Fatal(http.ListenAndServe(serviceAddr, mux))
}
