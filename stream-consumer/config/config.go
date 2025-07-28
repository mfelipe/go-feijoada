package config

import (
	_ "embed"
	"time"

	sbcfg "github.com/mfelipe/go-feijoada/stream-buffer/config"
	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	prefix = "SC"
)

//go:embed base.yaml
var baseCfg []byte

func Load() *Config {
	return utilscfg.Load[Config](prefix, baseCfg)
}

type Config struct {
	Log        utilslog.Config `json:"log" koanf:"log"`
	Consumer   Consumer        `json:"consumer" koanf:"consumer,required"`
	DynamoDB   DynamoDB        `json:"dynamoDB" koanf:"dynamoDB"`
	Repository sbcfg.Config    `json:"repository" koanf:"repository,required"`
}

type Consumer struct {
	BatchSize int           `json:"batchSize" koanf:"batchSize,required,gt=5"`
	Interval  time.Duration `json:"interval" koanf:"interval,required"`
}

type DynamoDB struct {
	Endpoint     string        `json:"endpoint" koanf:"endpoint"`
	TableName    string        `json:"tableName" koanf:"tableName,required"`
	RetryWaitMax time.Duration `json:"retryWaitMax" koanf:"retryWaitMax,required"`
	RetryMax     int           `json:"retryMax" koanf:"retryMax,required"`
}
