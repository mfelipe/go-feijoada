package config

import (
	_ "embed"
	"time"

	"github.com/rs/zerolog"

	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	prefix = "KP"
)

//go:embed base.yaml
var baseCfg []byte

func Load() *Consumer {
	return utilscfg.Load[Consumer](prefix, baseCfg)
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
