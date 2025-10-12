package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aleesa-misc-go/internal/log"
	"aleesa-misc-go/internal/misc"

	"github.com/go-redis/redis/v8"
)

func main() {
	var (
		err     error
		logfile *os.File
	)

	misc.ReadConfig()

	if misc.Config.Log != "" {
		logfile, err = os.OpenFile(misc.Config.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			log.Error(fmt.Sprintf("Unable to open log file %s: %s", misc.Config.Log, err))
			os.Exit(1)
		}
	} else {
		logfile = os.Stderr
	}

	log.Init(misc.Config.Loglevel, logfile)

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

	log.Info("Aleesa-misc-go started")

	for msg := range ch {
		if !misc.Shutdown {
			misc.MsgParser(misc.Ctx, msg.Payload)
		}
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
