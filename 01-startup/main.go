package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {

	// server
	server := &http.Server{
		Addr:         "0.0.0.0:9000",
		Handler:      http.HandlerFunc(helloworld),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	serverErrChan := make(chan error, 1)

	// start server in a goroutine
	go func() {
		log.Printf("starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("error: listening and serving: %s", err)
		}
	}()

	if err, ok := <-serverErrChan; ok && err != nil {
		log.Fatalf("error: server error: %s", err)
	}
}

// helloworld: basic http handler with response hello world string
func helloworld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World!")
}
