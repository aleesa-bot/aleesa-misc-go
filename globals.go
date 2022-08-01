package main

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

// Config - это у нас глобальная штука :)
var config myConfig

// To break circular message forwarding we must set some sane default, it can be overridden via config
var forwardMax int64 = 5

// Объектики клиента-редиски
var redisClient *redis.Client
var subscriber *redis.PubSub

// Main context
var ctx = context.Background()

// Ставится в true, если мы получили сигнал на выключение
var shutdown = false

// Канал, в который приходят уведомления для хэндлера сигналов от траппера сигналов
var sigChan = make(chan os.Signal, 1)

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
