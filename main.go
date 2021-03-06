package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// Производит некоторую инициализацию перед запуском main()
func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableQuote:           true,
		DisableLevelTruncation: false,
		DisableColors:          true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
	})

	readConfig()

	// no panic, no trace
	switch config.Loglevel {
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
}

// Основная функция программы, не добавить и не убавить
func main() {
	// Main context
	var ctx = context.Background()

	// Откроем лог и скормим его логгеру
	if config.Log != "" {
		logfile, err := os.OpenFile(config.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Unable to open log file %s: %s", config.Log, err)
		}

		log.SetOutput(logfile)
	}

	// Иницализируем клиента
	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Server, config.Port),
	})

	log.Debugf("Lazy connect() to redis at %s:%d", config.Server, config.Port)
	subscriber = redisClient.Subscribe(ctx, config.Channel)

	// Самое время поставить траппер сигналов
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go sigHandler()

	// Обработчик событий от редиски
	for {
		if shutdown {
			time.Sleep(1 * time.Second)
			continue
		}

		msg, err := subscriber.ReceiveMessage(ctx)

		if err != nil {
			if !shutdown {
				log.Warnf("Unable to connect to redis at %s:%d: %s", config.Server, config.Port, err)
			}

			time.Sleep(1 * time.Second)
			continue
		}

		go msgParser(ctx, msg.Payload)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
