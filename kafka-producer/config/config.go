package config

import (
	"embed"
	"time"

	"github.com/rs/zerolog"

	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	Prefix = "KP"
)

//go:embed base.yaml
var baseCfg embed.FS

func Load() *Consumer {
	var cfg Consumer
	utilscfg.Load(Prefix, baseCfg, &cfg)

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
