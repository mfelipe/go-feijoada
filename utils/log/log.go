package log

import "github.com/rs/zerolog"

func InitializeGlobal(cfg Config) {
	zerolog.SetGlobalLevel(cfg.Level)
}
