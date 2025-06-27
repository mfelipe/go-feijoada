package config

import (
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	delim = "."
)

func Load[T any](prefix string, filePath string, cfg *T) {
	k := koanf.New(delim)

	// Load YAML config file
	if err := k.Load(file.Provider(filePath), yaml.Parser()); err != nil {
		log.Panicf("error loading base yaml config: %v", err)
	}

	// Load Environment Variables config and merge
	err := k.Load(env.Provider(prefix, delim, func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", delim, -1)
	}), nil)
	if err != nil {
		log.Panicf("error loading config from environment variables: %v", err)
	}

	if err = k.Unmarshal(strings.ToLower(prefix), &cfg); err != nil {
		log.Panicf("error unmarshalling loaded configuration: %v", err)
	}
}
