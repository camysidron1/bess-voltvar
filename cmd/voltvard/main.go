package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/bess-voltvar/internal/config"
	"github.com/example/bess-voltvar/internal/controller"
	"github.com/example/bess-voltvar/internal/io"
	"github.com/example/bess-voltvar/internal/telem"
	"github.com/example/bess-voltvar/pkg/api"
)

func main() {
	cfgPath := flag.String("config", "./configs/site.example.yaml", "path to config yaml")
	addr := flag.String("addr", ":8080", "http listen address")
	tick := flag.Duration("tick", 100*time.Millisecond, "control loop tick")
	flag.Parse()

	telem.Setup() // basic logger only

	cfg, err := config.LoadFromFile(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// IO stubs (replace with real implementations in production)
	measProv := io.NewLocalMeasurements()
	pcsSink := io.NewLocalPCS()

	ctrl := controller.NewController(cfg, measProv, pcsSink, *tick)

	// API server
	srv := api.NewServer(ctrl, cfgPath)
	go func() {
		if err := srv.ListenAndServe(*addr); err != nil {
			log.Fatalf("api server error: %v", err)
		}
	}()

	// Control loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ctrl.Run(ctx)

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("shutting down...")
	cancel()
	time.Sleep(200 * time.Millisecond)
}
