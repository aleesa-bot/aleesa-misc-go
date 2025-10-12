package misc

import (
	"os"
	"syscall"

	"aleesa-misc-go/internal/log"
)

// SigHandler хэндлер сигналов закрывает все бд и сваливает из приложения.
func SigHandler() {
	var err error

	log.Info("Install signal handler")

	for {
		var s = <-SigChan
		switch s {
		case syscall.SIGINT:
			log.Info("Got SIGINT, quitting")
		case syscall.SIGTERM:
			log.Info("Got SIGTERM, quitting")
		case syscall.SIGQUIT:
			log.Info("Got SIGQUIT, quitting")

		// Заходим на новую итерацию, если у нас "неинтересный" сигнал.
		default:
			continue
		}

		// Чтобы не срать в логи ошибками от редиски, проставим shutdown state приложения в true.
		Shutdown = true

		// Отпишемся от всех каналов и закроем коннект к редиске.
		if err = Subscriber.Unsubscribe(Ctx); err != nil {
			log.Errorf("Unable to unsubscribe from redis channels cleanly: %s", err)
		}

		if err = Subscriber.Close(); err != nil {
			log.Errorf("Unable to close redis connection cleanly: %s", err)
		}

		os.Exit(0)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
