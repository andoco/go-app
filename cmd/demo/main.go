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
		fmt.Fprintf(w, "Ok at %v", time.Now())
	})
	app.AddHttp(mux)

	fmt.Println("Press CTRL+c to exit")
	app.Start()
}
