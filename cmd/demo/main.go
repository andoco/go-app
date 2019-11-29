package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andoco/go-app"
	"github.com/aws/aws-sdk-go/service/sqs"
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

	a.AddSQSWithConfig(&app.SQSWorkerConfig{Endpoint: "http://localhost:4576", ReceiveQueue: "http://localhost:4576/queue/test-queue"}, func(msg *sqs.Message) error {
		fmt.Printf("Handling message %v\n", msg)
		return nil
	})

	a.AddSQSWithConfig(&app.SQSWorkerConfig{Endpoint: "http://localhost:4576", ReceiveQueue: "http://localhost:4576/queue/test-queue-2"}, func(msg *sqs.Message) error {
		fmt.Printf("Handling message %v\n", msg)
		return nil
	})

	fmt.Println("Press CTRL+c to exit")
	a.Start()
}
