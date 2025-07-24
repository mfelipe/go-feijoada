package config

import (
	"embed"
	"errors"
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

const (
	delim = "."
)

func Load[T any](prefix string, baseConfig embed.FS, cfg *T) {
	k := koanf.New(delim)

	// Load YAML config file
	if err := k.Load(newMemoryProvider(baseConfig), yaml.Parser()); err != nil {
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

// memoryProvider implements a koanf provider for reading a configuration from memory, read via embed.FS
type memoryProvider struct {
	baseConfig embed.FS
}

// newMemoryProvider returns a memoryProvider provider.
func newMemoryProvider(baseConfig embed.FS) *memoryProvider {
	return &memoryProvider{baseConfig: baseConfig}
}

// ReadBytes reads the contents of a file on disk and returns the bytes.
// TODO: make it read the directory for multiple files (i.e. *.yaml)
func (m *memoryProvider) ReadBytes() ([]byte, error) {
	return m.baseConfig.ReadFile("base.yaml")
}

// Read is not supported by the file provider.
func (m *memoryProvider) Read() (map[string]interface{}, error) {
	return nil, errors.New("memoryProvider provider does not support this method")
}
