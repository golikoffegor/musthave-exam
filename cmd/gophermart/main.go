package main

import (
	"context"
	"musthave-exam/internal/app"
	"musthave-exam/internal/handler"
	"musthave-exam/internal/process"
	"musthave-exam/internal/repository"
	"musthave-exam/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	ap, err := app.NewApp()
	if err != nil {
		ap.Logger.Fatal(err)
	}
	defer ap.DB.Close()

	go handleSignals(ap.CncF, ap.Logger)

	//set repo
	repo := repository.NewRepository(ap.DB, ap.Logger)
	//set handlers
	vhandler := handler.NewHandler(repo, ap.Logger, ap.Flags)

	// ap.Logger.WithField("Flags", *ap.Flags).Info("App init")
	proc := process.NewOrderProcess(repo, ap.Logger, (*ap.Flags).RecalcAddress)
	proc.StartTransactionProcessing(ap.Ctx, repo)
	go proc.WaitTransactionProcessing(ap.Ctx)
	repo.SetProcess(proc.Listener)

	go func() {
		ap.Logger.Info("Start ListenAndServe")
		ap.Logger.Fatal(http.ListenAndServe(ap.Flags.Endpoint, router.InitRouter(*vhandler)))
	}()

	<-ap.Ctx.Done()
	ap.Logger.Info("Server stopped gracefully")
}

func handleSignals(cancel context.CancelFunc, logger *logrus.Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал
	sig := <-sigs
	logger.Info("Received signal: ", sig)
	cancel()
}
