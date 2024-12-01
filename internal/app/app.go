package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/SergeyBogomolovv/image-compressor/internal/config"
	"github.com/SergeyBogomolovv/image-compressor/internal/controller"
	"github.com/SergeyBogomolovv/image-compressor/internal/service"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config) *App {
	router := http.NewServeMux()
	service := service.New(log, cfg.OutputPath)
	controller.Register(router, service)

	return &App{
		server: &http.Server{
			Addr:    cfg.Addr,
			Handler: router,
		},
		log: log,
	}
}

func (a *App) Run() {
	const op = "app.Run"
	log := a.log.With(slog.String("op", op))

	go a.server.ListenAndServe()
	log.Info("application started", slog.String("addr", a.server.Addr))
}

func (a *App) Stop() {
	const op = "app.Stop"
	log := a.log.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a.server.Shutdown(ctx)

	log.Info("application stopped")
}
