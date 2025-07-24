package config

import (
	"embed"
	"time"

	"github.com/rs/zerolog"

	svcfg "github.com/mfelipe/go-feijoada/schema-validator/config"
	sbcfg "github.com/mfelipe/go-feijoada/stream-buffer/config"
	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	Prefix = "KC"
)

//go:embed base.yaml
var baseCfg embed.FS

func Load() *Consumer {
	var cfg Consumer
	utilscfg.Load(Prefix, baseCfg, &cfg)

	return &cfg
}

type Consumer struct {
	Log                         utilslog.Config `json:"log" koanf:"log"`
	SchemaValidator             svcfg.Config    `json:"schemaValidator" koanf:"schemaValidator,required"`
	Kafka                       Kafka           `json:"kafka" koanf:"kafka,required"`
	Repository                  sbcfg.Config    `json:"repository" koanf:"repository,required"`
	MaxProcessRoutines          int             `json:"maxProcessRoutines" koanf:"maxProcessRoutines,required,gt=0"`
	MaxPollRecords              int             `json:"maxPollRecords" koanf:"maxPollRecords,required,gt=0"`
	PartitionRecordsChannelSize int             `json:"partitionRecordsChannelSize" koanf:"partitionRecordsChannelSize,required,gte=5"`
	CloseTimeout                time.Duration   `json:"closeTimeout" koanf:"closeTimeout,required"`
}

type Kafka struct {
	Brokers string `json:"brokers" koanf:"brokers"`
	Group   string `json:"group" koanf:"group"`
	Topics  string `json:"topics" koanf:"topics"`
}

func (k Kafka) MarshalZerologObject(e *zerolog.Event) {
	e.Str("brokers", k.Brokers).
		Str("group", k.Group).
		Str("topics", k.Topics)
}
