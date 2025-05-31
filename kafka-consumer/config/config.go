package config

import (
	"log"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	sbcfg "github.com/mfelipe/go-feijoada/stream-buffer/config"
)

const (
	delim  = "."
	prefix = "kc_"
)

func Load() Consumer {
	k := koanf.New(".")

	// Load YAML config.
	if err := k.Load(file.Provider("base.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("error loading base yaml config: %v", err)
	}

	err := k.Load(env.Provider(prefix, delim, func(s string) string {
		return strings.Replace(strings.TrimPrefix(s, prefix), "_", delim, -1)
	}), nil)
	if err != nil {
		log.Fatalf("error loading config from environment variables: %v", err)
	}

	var c Consumer
	if err = k.Unmarshal("kc", &c); err != nil {
		log.Fatalf("error unmarshalling loaded configuration: %v", err)
	}

	return c
}

type Consumer struct {
	Kafka                       Kafka         `json:"kafka" koanf:"kafka,required"`
	Stream                      sbcfg.Config  `json:"stream" koanf:"stream,required"`
	MaxProcessRoutines          int           `json:"maxProcessRoutines" koanf:"maxProcessRoutines,required,gte=10"`
	PartitionRecordsChannelSize int           `json:"partitionRecordsChannelSize" koanf:"partitionRecordsChannelSize,required,gte=5"`
	CloseTimeout                time.Duration `json:"closeTimeout" koanf:"closeTimeout,required"`
}

type Kafka struct {
	Brokers string `json:"brokers" koanf:"brokers"`
	Group   string `json:"group" koanf:"group"`
	Topics  string `json:"topics" koanf:"topics"`
}
