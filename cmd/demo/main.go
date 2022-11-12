package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andoco/go-app"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	a := app.NewApp(app.NewAppConfig("FooDemo").WithMetrics("foo_demo"))

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

	fooProcessed := a.Metrics.NewCounterVec("foo_processed_total", "Total number of processed foos", []string{"bar"})

	msgRouter := app.NewMsgRouter()
	msgRouter.HandleFunc("foo", func(msg *app.MsgContext) error {
		msg.Logger.Info().Msg("Handling message")
		fooProcessed.With(prometheus.Labels{"bar": "baz"}).Inc()
		return nil
	})

	a.AddSQSWithConfig(&app.SQSWorkerConfig{Endpoint: "http://localhost:4566", ReceiveQueue: "http://localhost:4566/queue/test-queue", MsgTypeKey: "msgType"}, msgRouter)

	a.Start()
}
