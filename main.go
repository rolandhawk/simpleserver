package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	dd      *statsd.Client
	clients int32 = 0

	request = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_request",
		Help: "Number of request.",
	})

	latency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "http_latency_seconds",
		Help: "Request latency.",
	})

	activeclients = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "http_clients",
			Help: "Number of active clients",
		},
		gauge,
	)
)

func main() {
	client, err := statsd.New(os.Getenv("DATADOG_HOST")+":8125",
		statsd.WithNamespace("simpleserver.dd."),
	)
	if err != nil {
		panic(err)
	}
	dd = client

	prometheus.MustRegister(request)
	prometheus.MustRegister(latency)
	prometheus.MustRegister(activeclients)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", handler)

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("start server at :8080")
	log.Fatal(s.ListenAndServe())
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		elapse := time.Since(start).Seconds()
		request.Inc()
		dd.Incr("http_request", []string{}, 1)
		latency.Observe(elapse)
		dd.Histogram("http_latency_seconds", elapse, []string{}, 1)
	}()

	sleep := (90 * time.Millisecond) + (time.Duration(rand.Int63n(20)) * time.Millisecond)
	atomic.AddInt32(&clients, 1)
	dd.Gauge("http_clients", gauge(), []string{}, 1)
	time.Sleep(sleep)
	atomic.AddInt32(&clients, -1)
	dd.Gauge("http_clients", gauge(), []string{}, 1)

	fmt.Fprintf(w, "Hello...\n")
}

func gauge() float64 {
	return float64(atomic.LoadInt32(&clients))
}
