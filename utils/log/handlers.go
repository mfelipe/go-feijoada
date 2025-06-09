package log

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
)

func CustomLoggerRecovery() gin.OptionFunc {
	return func(e *gin.Engine) {
		recHandler := gin.CustomRecoveryWithWriter(log.Writer(), recoveryHandler)
		logHandler := gin.LoggerWithConfig(gin.LoggerConfig{
			Output: log.Writer(),
		})
		e.Use(logHandler, recHandler)
	}
}

func recoveryHandler(c *gin.Context, err any) {
	if err != nil {
		switch e := err.(type) {
		case error:
			zlog.Err(e).Msg("panic")
		default:
			zlog.Error().Msgf("panic: %v", err)
		}
	}

	if !c.IsAborted() {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
