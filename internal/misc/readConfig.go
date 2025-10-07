package misc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hjson/hjson-go"
)

// ReadConfig читает и валидирует конфиг, а также выставляет некоторые default-ы, если значений для параметров в конфиге
// нет.
func ReadConfig() {
	configLoaded := false
	executablePath, err := os.Executable()

	if err != nil {
		slog.Error(fmt.Sprintf("Unable to get current executable path: %s", err))
	}

	configJSONPath := fmt.Sprintf("%s/data/config.json", filepath.Dir(executablePath))

	locations := []string{
		"~/.aleesa-misc-go.json",
		"~/aleesa-misc-go.json",
		"/etc/aleesa-misc-go.json",
		configJSONPath,
	}

	for _, location := range locations {
		fileInfo, err := os.Stat(location)

		// Предполагаем, что файла либо нет, либо мы не можем его прочитать, второе надо бы логгировать, но пока забьём
		if err != nil {
			continue
		}

		// Конфиг-файл длинноват для конфига, попробуем следующего кандидата
		if fileInfo.Size() > 65535 {
			slog.Warn(fmt.Sprintf("Config file %s is too long for config, skipping", location))

			continue
		}

		buf, err := os.ReadFile(location)

		// Не удалось прочитать, попробуем следующего кандидата
		if err != nil {
			slog.Warn(fmt.Sprintf("Skip reading config file %s: %s", location, err))

			continue
		}

		// Исходя из документации, hjson какбы умеет парсить "кривой" json, но парсит его в map-ку.
		// Интереснее на выходе получить структурку: то есть мы вначале конфиг преобразуем в map-ку, затем эту map-ку
		// сериализуем в json, а потом json преврщааем в стркутурку. Не очень эффективно, но он и не часто требуется.
		var (
			sampleConfig myConfig
			tmp          map[string]interface{}
		)

		err = hjson.Unmarshal(buf, &tmp)

		// Не удалось распарсить - попробуем следующего кандидата
		if err != nil {
			slog.Warn(fmt.Sprintf("Skip parsing config file %s: %s", location, err))

			continue
		}

		tmpjson, err := json.Marshal(tmp)

		// Не удалось преобразовать map-ку в json
		if err != nil {
			slog.Warn(fmt.Sprintf("Skip parsing config file %s: %s", location, err))

			continue
		}

		if err := json.Unmarshal(tmpjson, &sampleConfig); err != nil {
			slog.Warn(fmt.Sprintf("Skip parsing config file %s: %s", location, err))

			continue
		}

		// Валидируем значения из конфига
		if sampleConfig.Server == "" {
			sampleConfig.Server = "localhost"
		}

		if sampleConfig.Port == 0 {
			sampleConfig.Port = 6379
		}

		if sampleConfig.Timeout == 0 {
			sampleConfig.Timeout = 10
		}

		if sampleConfig.Loglevel == "" {
			sampleConfig.Loglevel = "info"
		}

		// sampleConfig.Log = "" if not set

		if sampleConfig.Channel == "" {
			slog.Error(fmt.Sprintf("Channel field in config file %s must be set", location))
		}

		// Частичная проверка, ровно то, куда мы _точно_ щепрввляем сообщения исходя из бизнес-логики приложения
		if sampleConfig.ForwardChannels.Games == "" {
			sampleConfig.ForwardChannels.Games = "games"
		}

		if sampleConfig.ForwardChannels.Phrases == "" {
			sampleConfig.ForwardChannels.Phrases = "phrases"
		}

		if sampleConfig.ForwardChannels.Webapp == "" {
			sampleConfig.ForwardChannels.Webapp = "webapp"
		}

		if sampleConfig.ForwardChannels.WebappGo == "" {
			sampleConfig.ForwardChannels.WebappGo = "webapp-go"
		}

		if sampleConfig.ForwardChannels.Craniac == "" {
			sampleConfig.ForwardChannels.Craniac = "craniac"
		}

		if sampleConfig.Csign == "" {
			slog.Error(fmt.Sprintf("Csign field in config file %s must be set", location))
		}

		if sampleConfig.ForwardsMax == 0 {
			sampleConfig.ForwardsMax = ForwardMax
		}

		Config = sampleConfig
		configLoaded = true

		slog.Info(fmt.Sprintf("Using %s as config file", location))

		break
	}

	if !configLoaded {
		slog.Error("Config was not loaded! Refusing to start.")
		os.Exit(1)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
