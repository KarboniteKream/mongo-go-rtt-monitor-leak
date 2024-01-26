package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

const URI = "mongodb+srv://FOO:BAR@EXAMPLE.COM"

func main() {
	if err := execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute() error {
	target, err := NewPinger(URI)
	if err != nil {
		return err
	}
	defer target.Close()

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Reconnecting...")
		err := target.Reconnect()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	srv := &http.Server{Addr: ":19508"}
	srvc := make(chan error)
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Listening on http://localhost:19508/ping")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			srvc <- err
		}
	}()

	for {
		select {
		case <-term:
			log.Print("Received SIGTERM")
			return nil
		case err := <-srvc:
			return err
		}
	}
}
