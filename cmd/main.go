package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"rsclabs-test/internal/repository/inmemorystorage"
	"syscall"
	"time"

	"rsclabs-test/config"
	"rsclabs-test/internal/controller/http"
	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/service"
	"rsclabs-test/internal/worker"
	"rsclabs-test/pkg/httpserver"
	"rsclabs-test/pkg/observe"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cnf := config.NewConfig()

	l := observe.NewZapLogger(cnf.AppName, os.Stdout)

	server := httpserver.InitFiberServer(cnf.AppName)

	inMemoryStorage := inmemorystorage.NewInMemoryStorage(cnf.MaxBanners, l)

	bannerRepository, err := repository.NewBannerRepository(inMemoryStorage)
	if err != nil {
		l.Fatal("failed to create banner repository", map[string]any{"err": err})
	}

	statisticsService := service.NewStatisticsService(
		bannerRepository,
		server,
		l,
	)

	statisticsWorker := worker.NewStatisticsWorker(
		bannerRepository,
		statisticsService,
		l,
	)

	go statisticsWorker.Run(ctx)

	http.NewRouter(
		bannerRepository,
		statisticsService,
		server,
		l,
	)

	go func() {
		if err := server.Listen(":" + cnf.Port); err != nil {
			l.Fatal("cannot run the server", map[string]any{"err": err})
		}
	}()

	l.Info("application started successfully", map[string]any{"port": cnf.Port})

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		l.Warning("stopping application services")
		signal.Stop(sigCh)
		close(sigCh)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		_ = server.ShutdownWithContext(shutdownCtx)
		_ = l.Stop()
		cancel()
	}()

	select {
	case <-sigCh:
		fmt.Println("received shutdown signal")
	case <-ctx.Done():
		fmt.Println("context cancelled")
	}
}
