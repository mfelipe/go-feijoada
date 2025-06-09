package config

import (
	"path/filepath"
	"time"

	svcfg "github.com/mfelipe/go-feijoada/schema-validator/config"
	sbcfg "github.com/mfelipe/go-feijoada/stream-buffer/config"
	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	Prefix = "KC"
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
	Log                         utilslog.Config `json:"log" koanf:"log"`
	SchemaValidator             svcfg.Config    `json:"schemaValidator" koanf:"schemaValidator,required"`
	Kafka                       Kafka           `json:"kafka" koanf:"kafka,required"`
	Stream                      sbcfg.Config    `json:"stream" koanf:"stream,required"`
	MaxProcessRoutines          int             `json:"maxProcessRoutines" koanf:"maxProcessRoutines,required,gt=0"`
	MaxPoolRecords              int             `json:"maxPoolRecords" koanf:"maxPoolRecords,required,gt=0"`
	PartitionRecordsChannelSize int             `json:"partitionRecordsChannelSize" koanf:"partitionRecordsChannelSize,required,gte=5"`
	CloseTimeout                time.Duration   `json:"closeTimeout" koanf:"closeTimeout,required"`
}

type Kafka struct {
	Brokers string `json:"brokers" koanf:"brokers"`
	Group   string `json:"group" koanf:"group"`
	Topics  string `json:"topics" koanf:"topics"`
}
