package main

// A naive implementation of http client.

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"jaeger/pkg/trace"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	reportAddr  string
	serviceName string
	requestAddr string
)

func init() {
	flag.StringVar(&reportAddr, "report_addr", "127.0.0.1:6831", "the address of jaeger agent.")
	flag.StringVar(&serviceName, "service_name", "client", "the name of service.")
	flag.StringVar(&requestAddr, "request_addr", "http://localhost:8181/", "")
	flag.Parse()
}

func request(tracer opentracing.Tracer, span opentracing.Span) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Set span into context.
	ctx = opentracing.ContextWithSpan(ctx, span)

	req, err := http.NewRequestWithContext(ctx, "GET", requestAddr, nil)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogKV("error", err)
		log.Println(err)
		return
	}

	span.LogKV("request", req.URL)
	ext.SpanKindRPCClient.Set(span)
	ext.Component.Set(span, "client")
	ext.HTTPUrl.Set(span, requestAddr)
	ext.HTTPMethod.Set(span, "GET")

	client := &http.Client{Timeout: time.Second * 10}

	// Inject span context into http headers.
	tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	resp, err := client.Do(req)
	if err != nil {
		span.LogKV("error", err)
		ext.Error.Set(span, true)
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogKV("error", err)
		log.Println(err)
		return
	}

	log.Println(string(ret))
	return
}

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	// Create a root span.
	span := tracer.StartSpan("start")
	defer span.Finish()

	// Add baggage items into span.
	span.Context().ForeachBaggageItem(func(k, v string) bool {
		span.LogKV(k, v)
		return true
	})

	request(tracer, span)
}
