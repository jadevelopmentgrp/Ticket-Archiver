package main

import (
	"context"
	"time"

	"github.com/jadevelopmentgrp/Tickets-Archiver/pkg/config"
	"github.com/jadevelopmentgrp/Tickets-Archiver/pkg/http"
	"github.com/jadevelopmentgrp/Tickets-Archiver/pkg/repository"
	"github.com/jadevelopmentgrp/Tickets-Archiver/pkg/s3client"
	"go.uber.org/zap"
)

func main() {
	conf := config.Parse[config.Config]()

	var logger *zap.Logger
	var err error
	if conf.ProductionMode {
		logger, err = zap.NewProduction(
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
		)
	} else {
		logger, err = zap.NewDevelopment(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	}

	if err != nil {
		panic(err)
	}

	logger.Info("Connecting to database...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := repository.ConnectPostgres(ctx, conf)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("Connected.")

	logger.Debug("Starting S3 client manager...")
	clientManager := s3client.NewShardedClientManager(conf, store)
	if err := clientManager.Load(ctx); err != nil {
		logger.Fatal("Failed to load S3 clients", zap.Error(err))
	}

	logger.Debug("Starting HTTP server...")

	server := http.NewServer(logger, conf, store, clientManager)
	go server.RemoveQueue.StartReaper()
	server.RegisterRoutes()
	server.Start()
}
