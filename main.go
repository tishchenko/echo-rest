package main

// http://www.gorillatoolkit.org/pkg/mux

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", simpleHandler).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD")
	r.HandleFunc("/echo", echoHandler).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD")

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	io.WriteString(setupHeader(w), `{"alive": true}`)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	io.WriteString(setupHeader(w), bodyString)
}

func setupHeader(w http.ResponseWriter) http.ResponseWriter {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	return w
}

func logRequest(r *http.Request) {
	log.Println()
	log.Println(r.URL.String())
	log.Println(r.Method)
	for k, v := range r.Header {
		log.Printf("%q: %q\n", k, v)
	}
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	log.Println(bodyString)
}
