package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aleesa-misc-go/internal/misc"

	"github.com/go-redis/redis/v8"
)

func main() {
	var (
		err      error
		loglevel slog.Level
	)

	misc.ReadConfig()

	loghandler := os.Stderr

	if misc.Config.Log != "" {
		loghandler, err = os.OpenFile(misc.Config.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			slog.Error(fmt.Sprintf("Unable to open log file %s: %s", misc.Config.Log, err))
			os.Exit(1)
		}
	}

	// no panic, no trace.
	switch misc.Config.Loglevel {
	case "error":
		loglevel = slog.LevelError

	case "warn":
		loglevel = slog.LevelWarn

	case "info":
		loglevel = slog.LevelInfo

	case "debug":
		loglevel = slog.LevelDebug

	default:
		loglevel = slog.LevelInfo

	}

	opts := &slog.HandlerOptions{
		// Use the ReplaceAttr function on the handler options
		// to be able to replace any single attribute in the log output
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// check that we are handling the time key
			if a.Key != slog.TimeKey {
				return a
			}

			t := a.Value.Time()

			// change the value from a time.Time to a String
			// where the string has the correct time format.
			a.Value = slog.StringValue(t.Format(time.DateTime))

			return a
		},

		Level: loglevel,
	}

	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				loghandler,
				opts,
			),
		),
	)

	// Иницализируем клиента Редиски.
	misc.RedisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", misc.Config.Server, misc.Config.Port),
	}).WithContext(misc.Ctx).WithTimeout(time.Duration(misc.Config.Timeout) * time.Second)

	// Обозначим, что хотим после соединения подписаться на события из канала config.Channel.
	misc.Subscriber = misc.RedisClient.Subscribe(misc.Ctx, misc.Config.Channel)

	// Самое время поставить траппер сигналов.
	signal.Notify(misc.SigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go misc.SigHandler()

	// Начнём выгребать события из редиски (длина конвеера/буфера канала по-умолчанию - 100 сообщений).
	ch := misc.Subscriber.Channel()

	slog.Info("Aleesa-misc-go started")

	for msg := range ch {
		if !misc.Shutdown {
			misc.MsgParser(misc.Ctx, msg.Payload)
		}
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
