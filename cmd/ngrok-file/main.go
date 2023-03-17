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

	"go.uber.org/zap"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	zapa "golang.ngrok.com/ngrok/log/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	flag.Parse()

	file := flag.Arg(0)

	zcfg := zap.NewDevelopmentConfig()
	zcfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)

	zlog, err := zcfg.Build()
	if err != nil {
		panic(err)
	}

	handler := newHandler(file)

	ctx := context.TODO()
	tunnel, err := ngrok.Listen(ctx,
		config.HTTPEndpoint(config.WithCompression()),
		ngrok.WithAuthtokenFromEnv(),
		ngrok.WithLogger(zapa.NewLogger(zlog)))

	if err != nil {
		panic(err)
	}

	fmt.Println("Visit", tunnel.URL())

	eg, ectx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ctx, cancel := context.WithCancel(ectx)

		var err error
		srv := http.Server{Handler: handler}
		go func() {
			defer cancel()
			err = srv.Serve(tunnel)
		}()

		<-ctx.Done()
		tunnel.Close()
		return err
	})

	eg.Go(func() error {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
		sig := <-exit
		return errors.New(sig.String())
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}

func newHandler(f string) http.Handler {
	if f == "" {
		return http.FileServer(http.Dir("."))
	}

	fstat, err := os.Stat(f)
	if err != nil {
		panic(err)
	}

	if fstat.IsDir() {
		return http.FileServer(http.Dir(f))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, f)
	})
}
