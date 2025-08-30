package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"aleesa-misc-go/internal/misc"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableQuote:           true,
		DisableLevelTruncation: false,
		DisableColors:          true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
	})

	misc.ReadConfig()

	// no panic, no trace.
	switch misc.Config.Loglevel {
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Иницализируем клиента Редиски.
	misc.RedisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", misc.Config.Server, misc.Config.Port),
	}).WithContext(misc.Ctx).WithTimeout(time.Duration(misc.Config.Timeout) * time.Second)

	// Обозначим, что хотим после соединения подписаться на события из канала config.Channel.
	misc.Subscriber = misc.RedisClient.Subscribe(misc.Ctx, misc.Config.Channel)

	// Откроем лог и скормим его логгеру.
	if misc.Config.Log != "" {
		logfile, err := os.OpenFile(misc.Config.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Unable to open log file %s: %s", misc.Config.Log, err)
		}

		log.SetOutput(logfile)
	}

	// Самое время поставить траппер сигналов.
	signal.Notify(misc.SigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go misc.SigHandler()

	// Начнём выгребать события из редиски (длина конвеера/буфера канала по-умолчанию - 100 сообщений).
	ch := misc.Subscriber.Channel()

	log.Infoln("Aleesa-misc-go started")

	for msg := range ch {
		if !misc.Shutdown {
			misc.MsgParser(misc.Ctx, msg.Payload)
		}
	}
}
