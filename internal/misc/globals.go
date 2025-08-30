package misc

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

// Config - это у нас глобальная штука.
var Config myConfig

// To break circular message forwarding we must set some sane default, it can be overridden via config.
var ForwardMax int64 = 5

// Объектики клиента-редиски.
var RedisClient *redis.Client
var Subscriber *redis.PubSub

// Ставится в true, если мы получили сигнал на выключение.
var Shutdown = false

// Канал, в который приходят уведомления для хэндлера сигналов от траппера сигналов.
var SigChan = make(chan os.Signal, 1)

// Main context.
var Ctx = context.Background()

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
