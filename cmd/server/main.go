// Program server shows how verbose logging can be enabled
// without program restart using ops HTTP endpoint.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v3"

	"github.com/marselester/diagnostic-kit"
)

var errInterrupt = errors.New("program interrupted")

func main() {
	// By default an exit code is set to indicate a failure since
	// there are more failure scenarios to begin with.
	exitCode := 1
	defer func() { os.Exit(exitCode) }()

	baseLogger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	baseLogger = log.With(baseLogger, "ts", log.DefaultTimestampUTC)
	standardLogger := level.NewFilter(baseLogger, level.AllowError())
	logger := &log.SwapLogger{}
	logger.Swap(baseLogger)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		apiAddr = fs.String("api-addr", ":8000", "HTTP service address")
		opsAddr = fs.String("ops-addr", ":9000", "HTTP ops address to expose metrics")
		debug   = fs.Bool("debug", false, "log debug information")
		_       = fs.String("config", "", "config file")
	)
	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("SERVICE"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)
	if err != nil {
		level.Error(logger).Log("msg", "parse flags", "err", err)
		return
	}

	if !*debug {
		logger.Swap(standardLogger)
	}
	http.DefaultServeMux.HandleFunc("/logging", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("form parse failed"))
		}

		switch r.Form.Get("logger") {
		case "standard":
			logger.Swap(standardLogger)
			w.Write([]byte("standard logger enabled"))
		case "debug":
			k, v := r.Form.Get("key"), r.Form.Get("value")
			if k == "" || v == "" {
				logger.Swap(baseLogger)
				w.Write([]byte("debug logger enabled"))
				return
			}

			logger.Swap(&diagnostic.FilterLogger{
				Hit:   baseLogger,
				Miss:  standardLogger,
				Key:   k,
				Value: v,
			})
			fmt.Fprintf(w, "filtered debug logger enabled: %s=%s", k, v)
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unknown logger"))
		}
	})

	opsServer := http.Server{
		Addr:    *opsAddr,
		Handler: http.DefaultServeMux,
	}
	apiServer := http.Server{
		Addr:         *apiAddr,
		Handler:      newAPIHandler(log.With(logger, "component", "api")),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var g run.Group
	{
		g.Add(func() error {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-sig:
				cancel()
				return errInterrupt
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func(err error) {
			level.Debug(logger).Log("actor", "signal", "msg", "interrupted")
			cancel()
		})
	}
	{
		g.Add(func() error {
			level.Debug(logger).Log("actor", "api", "msg", "starting", "addr", *apiAddr)
			return apiServer.ListenAndServe()
		}, func(err error) {
			level.Debug(logger).Log("actor", "api", "msg", "interrupted")
			err = apiServer.Shutdown(ctx)
			level.Debug(logger).Log("actor", "api", "msg", "shutdown", "err", err)
		})
	}
	{
		g.Add(func() error {
			level.Debug(logger).Log("actor", "ops", "msg", "starting", "addr", *opsAddr)
			return opsServer.ListenAndServe()
		}, func(err error) {
			level.Debug(logger).Log("actor", "ops", "msg", "interrupted")
			err = opsServer.Shutdown(ctx)
			level.Debug(logger).Log("actor", "ops", "msg", "shutdown", "err", err)
		})
	}
	err = g.Run()

	// The program terminates successfully if it received INT/TERM signal.
	if err == errInterrupt {
		exitCode = 0
		level.Debug(logger).Log("actor", "all", "msg", "stopped", "err", err)
	} else {
		level.Error(logger).Log("actor", "all", "msg", "stopped", "err", err)
	}
}

func newAPIHandler(logger log.Logger) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		level.Debug(logger).Log("user_id", userID, "path", r.URL.Path)
		level.Info(logger).Log("user_id", userID, "path", r.URL.Path)
		level.Warn(logger).Log("user_id", userID, "path", r.URL.Path)
		level.Error(logger).Log("user_id", userID, "path", r.URL.Path)
	})

	return r
}
