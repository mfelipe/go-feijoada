package log

import "github.com/rs/zerolog"

type Config struct {
	Level zerolog.Level `json:"level" koanf:"level"`
}
