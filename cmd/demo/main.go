package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andoco/go-app"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

func main() {
	a := app.NewApp("Demo")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok(1) at %v", time.Now())
	})
	a.AddHttp(mux, 8081)

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok(2) at %v", time.Now())
	})
	a.AddHttp(mux2, 8082)

	a.AddPrometheus()

	opsProcessed := promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})

	opsProcessed.Inc()

	a.AddSQSWithConfig(&app.SQSWorkerConfig{Endpoint: "http://localhost:4576", ReceiveQueue: "http://localhost:4576/queue/test-queue"}, func(msg *sqs.Message, logger zerolog.Logger) error {
		logger.Info().Msg("Handling message")
		return nil
	})

	a.AddSQSWithConfig(&app.SQSWorkerConfig{Endpoint: "http://localhost:4576", ReceiveQueue: "http://localhost:4576/queue/test-queue-2"}, func(msg *sqs.Message, logger zerolog.Logger) error {
		logger.Info().Msg("Handling message")
		return nil
	})

	fmt.Println("Press CTRL+c to exit")
	a.Start()
}
