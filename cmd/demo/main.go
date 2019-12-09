package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andoco/go-app"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

	a.AddPrometheus("/metrics", 9090)

	opsProcessed := promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})

	opsProcessed.Inc()

	msgRouter := app.NewMsgRouter()
	msgRouter.HandleFunc("foo", func(msg *app.MsgContext) error {
		msg.Logger.Info().Msg("Handling message")
		return nil
	})

	a.AddSQSWithConfig(&app.SQSWorkerConfig{Endpoint: "http://localhost:4576", ReceiveQueue: "http://localhost:4576/queue/test-queue"}, msgRouter)

	a.Start()
}
