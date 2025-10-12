package misc

import (
	"context"
	"encoding/json"
	"strings"

	"aleesa-misc-go/internal/log"
)

// MsgParser горутинка, которая парсит json-чики прилетевшие из REDIS-ки.
func MsgParser(ctx context.Context, msg string) {
	var (
		sendTo = "craniac"
		j      rMsg
	)

	log.Debugf("Incomming raw json: %s", msg)

	if err := json.Unmarshal([]byte(msg), &j); err != nil {
		log.Warnf("Unable to to parse message from redis channel: %s", err)

		return
	}

	// Validate our j.
	if exist := j.From; exist == "" {
		log.Warnf("Incorrect msg from redis, no from field: %s", msg)

		return
	}

	if exist := j.Chatid; exist == "" {
		log.Warnf("Incorrect msg from redis, no chatid field: %s", msg)

		return
	}

	if exist := j.Userid; exist == "" {
		log.Warnf("Incorrect msg from redis, no userid field: %s", msg)

		return
	}

	// j.Threadid может быть пустым, значит либо нам его не дали, либо дали пустым. Это нормально.

	if exist := j.Message; exist == "" {
		log.Warnf("Incorrect msg from redis, no message field: %s", msg)

		return
	}

	if exist := j.Plugin; exist == "" {
		log.Warnf("Incorrect msg from redis, no plugin field: %s", msg)

		return
	}

	if exist := j.Mode; exist == "" {
		log.Warnf("Incorrect msg from redis, no mode field: %s", msg)

		return
	}

	// j.Misc.Answer может и не быть, тогда ответа на такое сообщение не будет.
	// j.Misc.Botnick тоже можно не передавать, тогда будет записана пустая строка.
	// j.Misc.Csign если нам его не передали, возьмём значение из конфига.
	if exist := j.Misc.Csign; exist == "" {
		j.Misc.Csign = Config.Csign
	}

	// j.Misc.Fwdcnt если нам его не передали, то будет 0.
	if exist := j.Misc.Fwdcnt; exist == 0 {
		j.Misc.Fwdcnt = 1
	}

	// j.Misc.GoodMorning может быть 1 или 0, по-умолчанию 0.
	// j.Misc.Msgformat может быть 1 или 0, по-умолчанию 0.
	// j.Misc.Username можно не передавать, тогда будет пустая строка.

	// Отвалидировались, теперь вернёмся к нашим баранам.

	// Если у нас циклическая пересылка сообщения, попробуем её тут разорвать, отбросив сообщение.
	if j.Misc.Fwdcnt > Config.ForwardsMax {
		log.Warnf("Discarding msg with fwd_cnt exceeding max value: %s", msg)

		return
	}

	j.Misc.Fwdcnt++

	// Классифицирем входящие сообщения. Первым делом, попробуем определить команды.
	if j.Message[0:len(j.Misc.Csign)] == j.Misc.Csign {
		// Может быть, это команда модуля phrases?
		var (
			done = false
			cmd  = j.Message[len(j.Misc.Csign):]
		)

		cmds := []string{"friday", "пятница", "proverb", "пословица", "пословиться", "fortune", "фортунка", "f", "ф",
			"karma", "карма", "rum", "ром", "vodka", "водка", "beer", "пиво", "tequila", "текила", "whisky", "виски",
			"absinthe", "абсент", "fuck"}

		for _, command := range cmds {
			if cmd == command {
				sendTo = Config.ForwardChannels.Phrases

				// Костыль для кармы.
				if cmd == "karma" || cmd == "карма" {
					j.Misc.Answer = 1
				}

				done = true

				break
			}
		}

		// Не угадали? акей, как насчёт модуля webapp-go?
		if !done {
			cmds = []string{"frog", "лягушка", "horse", "лошадь", "лошадка", "rabbit", "bunny", "кролик",
				"snail", "улитка", "cat", "кис", "fox", "лис", "buni", "anek", "анек", "анекдот",
				"xkcd", "monkeyuser", "tits", "boobs", "tities", "boobies", "сиси", "сисечки", "butt",
				"booty", "ass", "попа", "попка", "drink", "праздник", "owl", "сова", "сыч", "w", "п", "погода",
				"погодка", "погадка"}

			for _, command := range cmds {
				if cmd == command {
					sendTo = Config.ForwardChannels.WebappGo
					done = true

					break
				}
			}
		}

		// Не фортануло? может, повезёт с модулем games?
		if !done {
			cmds = []string{"dig", "копать", "fish", "fishing", "рыба", "рыбка", "рыбалка"}

			for _, command := range cmds {
				if cmd == command {
					sendTo = Config.ForwardChannels.Games
					done = true

					break
				}
			}
		}

		// Нет? Вдруг это комплексная команда модуля webapp?
		if !done {
			cmdLen := len(cmd)

			cmds := []string{"w ", "п ", "погода ", "погодка ", "погадка ", "weather "}

			for _, command := range cmds {
				if cmdLen > len(command) && cmd[0:len(command)] == command {
					sendTo = Config.ForwardChannels.WebappGo
					done = true

					break
				}
			}
		}

		// Опять мимо? Давай тогда попытаем удачу в поиске комплексной команды для phrases.
		if !done {
			cmdLen := len(cmd)

			cmds := []string{"karma ", "карма ", "rum ", "ром ", "vodka ", "водка ", "beer ", "пиво ", "tequila ",
				"текила ", "whisky ", "виски ", "absinthe ", "абсент "}

			for _, command := range cmds {
				if cmdLen > len(command) && cmd[0:len(command)] == command {
					sendTo = Config.ForwardChannels.Phrases

					if command == "karma " || command == "карма " {
						// Костыль для кармы
						j.Misc.Answer = 1
					}

					break
				}
			}
		}
	} else {
		// Попробуем выискать изменение кармы.
		msgLen := len(j.Message)

		// ++ или -- на конце фразы - это наш кандидат.
		if msgLen > len("++") {
			if j.Message[msgLen-len("--"):msgLen] == "--" || j.Message[msgLen-len("++"):msgLen] == "++" {
				// Предполагается, что менять карму мы будем для одной фразы, то есть для 1 строки
				if len(strings.Split(j.Message, "\n")) == 1 {
					sendTo = Config.ForwardChannels.Phrases

					// Костыль для кармы
					j.Misc.Answer = 1
				}
			}
		}
	}

	// Настало время формировать json и засылать его в дальше.
	var message sMsg
	message.From = j.From
	message.Userid = j.Userid
	message.Chatid = j.Chatid
	message.Threadid = j.Threadid
	message.Message = j.Message
	message.Plugin = j.Plugin
	message.Mode = j.Mode
	message.Misc.Fwdcnt = j.Misc.Fwdcnt
	message.Misc.Csign = j.Misc.Csign
	message.Misc.Username = j.Misc.Username
	message.Misc.Answer = j.Misc.Answer
	message.Misc.Botnick = j.Misc.Botnick
	message.Misc.Msgformat = j.Misc.Msgformat
	message.Misc.GoodMorning = j.Misc.GoodMorning

	data, err := json.Marshal(message)

	if err != nil {
		log.Warnf("Unable to to serialize message for redis: %s", err)

		return
	}

	// Заталкиваем наш json в редиску.
	if err := RedisClient.Publish(ctx, sendTo, data).Err(); err != nil {
		log.Warnf("Unable to send data to redis channel %s: %s", sendTo, err)
	} else {
		log.Debugf("Send msg to redis channel %s: %s", sendTo, string(data))
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
