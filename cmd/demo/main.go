package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andoco/go-app"
)

func main() {
	app := app.NewApp()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok(1) at %v", time.Now())
	})
	app.AddHttp(mux, 8081)

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok(2) at %v", time.Now())
	})
	app.AddHttp(mux2, 8082)

	fmt.Println("Press CTRL+c to exit")
	app.Start()
}
