package config

import (
	"github.com/knadh/koanf/providers/rawbytes"
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

const (
	delim    = "."
	oldDelim = "_"
)

func Load[T any](prefix string, cfgContent []byte) *T {
	k := koanf.New(delim)
	cfg := new(T)

	if err := k.Load(rawbytes.Provider(cfgContent), yaml.Parser()); err != nil {
		log.Panicf("error loading base configuration file: %v", err)
	}

	// Load Environment Variables config and merge
	if err := k.Load(env.Provider(prefix, delim, func(s string) string {
		return strings.Replace(strings.ToLower(s), oldDelim, delim, -1)
	}), nil); err != nil {
		log.Panicf("error loading config from environment variables: %v", err)
	}

	if err := k.Unmarshal(strings.ToLower(prefix), cfg); err != nil {
		log.Panicf("error unmarshalling loaded configuration: %v", err)
	}

	return cfg
}
