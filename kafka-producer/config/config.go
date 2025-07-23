package config

import (
	"path/filepath"
	"time"

	"github.com/rs/zerolog"

	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	Prefix = "KP"
)

func Load() *Consumer {
	path, err := filepath.Abs("../config/base.yaml")
	if err != nil {
		panic(err)
	}
	var cfg Consumer
	utilscfg.Load[Consumer](Prefix, path, &cfg)

	return &cfg
}

type Consumer struct {
	Log          utilslog.Config `json:"log" koanf:"log"`
	Kafka        Kafka           `json:"kafka" koanf:"kafka,required"`
	CloseTimeout time.Duration   `json:"closeTimeout" koanf:"closeTimeout,required"`
}

type Kafka struct {
	Brokers string `json:"brokers" koanf:"brokers"`
}

func (k Kafka) MarshalZerologObject(e *zerolog.Event) {
	e.Str("brokers", k.Brokers)
}
