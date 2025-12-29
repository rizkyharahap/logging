package main

import (
	"context"
	"fmt"
	"logging-challenge/middlewares"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		oscall := <-ch
		log.Warn().Msgf("system call:%+v", oscall)
		cancel()
	}()

	r := mux.NewRouter()
	r.HandleFunc("/", handler)

	// start: set up any of your logger configuration here if necessary

	// Set middleware
	r.Use(middlewares.Logging)

	// handle log level by environment variable
	if os.Getenv("LOG_LEVEL") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// open log file
	lf, err := os.OpenFile(
		"logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open log file")
	}
	defer lf.Close()

	// Set multiwriters for stdout and log file
	multiWriters := zerolog.MultiLevelWriter(os.Stdout, lf)

	log.Logger = zerolog.New(multiWriters).With().Timestamp().Logger()
	// end: set up any of your logger configuration here

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to listen and serve http server")
		}
	}()
	<-ctx.Done()

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("failed to shutdown http server gracefully")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// creating a new context from the logger instance
	log := log.Ctx(ctx).With().Str("func", "handler").Logger()
	log.Debug().Msg("processing request")

	name := r.URL.Query().Get("name")

	res, err := greeting(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Str("response", res).Msg("request processed")

	w.Write([]byte(res))
}

func greeting(ctx context.Context, name string) (string, error) {
	log := log.Ctx(ctx)
	log.Debug().Str("func", "greeting").Str("name", name).Msg("processing greeting")

	if len(name) < 5 {
		return fmt.Sprintf("Hello %s! Your name is to short\n", name), nil
	}
	return fmt.Sprintf("Hi %s", name), nil
}
