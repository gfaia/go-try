package main

import (
	"flag"
	"log"

	"jaeger/pkg/trace"

	"github.com/gin-gonic/gin"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
)

var (
	reportAddr  string
	serviceName string
	serviceAddr string
)

func init() {
	flag.StringVar(&reportAddr, "report_addr", "127.0.0.1:6831", "the address of jaeger agent.")
	flag.StringVar(&serviceName, "service_name", "server", "")
	flag.StringVar(&serviceAddr, "service_addr", ":8181", "")
	flag.Parse()
}

func main() {
	tracer, closer, err := trace.NewJaegerTracer(serviceName, reportAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	r := gin.New()
	r.Use(ginhttp.Middleware(tracer))
	r.GET("/", func(c *gin.Context) {
		span := opentracing.SpanFromContext(c.Request.Context())
		if span == nil {
			log.Fatalf("span is nill")
		}
		c.String(200, "test")
	})
	r.Run(serviceAddr)
}
