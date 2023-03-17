package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	zapa "golang.ngrok.com/ngrok/log/zap"
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

	ctx := context.TODO()
	tunnel, err := ngrok.Listen(ctx,
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
		ngrok.WithLogger(zapa.NewLogger(zlog)))

	if err != nil {
		panic(err)
	}

	if file == "" {
		fmt.Println("WARN: exposing current directory")
	}

	fmt.Println("Visit", tunnel.URL())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	})

	if err := http.Serve(tunnel, handler); err != nil {
		panic(err)
	}
}
