package main

import (
	"context"
	"log"
	"mailer-ms/config"
	"mailer-ms/mail"
	"mailer-ms/queue"
	"mailer-ms/tracer"
	"os"
	"os/signal"
	"syscall"
)

// The version/build, this gets replaced at build time to the commit SHA
// with the use of linker flags. see ldfflags on the makefile build cmd

var version = "development"
var build = "development"

func init() {
	log.Println("[ GIT ] build:   ", build)
	log.Println("[ GIT ] version: ", version)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("[CONFIG] failed to parse config: %v", err)
	}

	err = tracer.Start(&cfg.Tracer)
	if err != nil {
		log.Fatalf("[TRACER] failed to init tracer: %v", err)
	}
	defer tracer.Stop(ctx)

	queue := queue.New(cfg.Rmq)
	mailer := mail.New(cfg, &queue)

	queue.ConsumerFn = mailer.HandleMailRequestDelivery

	queue.Start()
	defer queue.Stop()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-exit
}
